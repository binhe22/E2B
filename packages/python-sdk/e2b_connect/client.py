import gzip
import json
import struct

from httpcore import ConnectionPool, AsyncConnectionPool, Response
from enum import Flag, Enum
from typing import Optional, Dict
from google.protobuf import json_format


class EnvelopeFlags(Flag):
    compressed = 0b00000001
    end_stream = 0b00000010


class Code(Enum):
    canceled = "canceled"
    unknown = "unknown"
    invalid_argument = "invalid_argument"
    deadline_exceeded = "deadline_exceeded"
    not_found = "not_found"
    already_exists = "already_exists"
    permission_denied = "permission_denied"
    resource_exhausted = "resource_exhausted"
    failed_precondition = "failed_precondition"
    aborted = "aborted"
    out_of_range = "out_of_range"
    unimplemented = "unimplemented"
    internal = "internal"
    unavailable = "unavailable"
    data_loss = "data_loss"
    unauthenticated = "unauthenticated"


class ConnectException(Exception):
    def __init__(self, status: Code, message: str):
        self.status = status
        self.message = message


envelope_header_length = 5
envelope_header_pack = ">BI"


def encode_envelope(*, flags: EnvelopeFlags, data):
    return encode_envelope_header(flags=flags.value, data=data) + data


def encode_envelope_header(*, flags, data):
    return struct.pack(envelope_header_pack, flags, len(data))


def decode_envelope_header(header):
    flags, data_len = struct.unpack(envelope_header_pack, header)
    return EnvelopeFlags(flags), data_len


def error_for_response(http_resp: Response):
    try:
        error = json.loads(http_resp.content)
    except (json.decoder.JSONDecodeError, KeyError):
        if http_resp.status == 429:
            return ConnectException(
                Code.resource_exhausted,
                f"{http_resp.content.decode()} The requests are being rate limited.",
            )
        elif http_resp.status == 502:
            return ConnectException(
                Code.unavailable,
                http_resp.content.decode(),
            )
        else:
            return ConnectException(
                Code.unknown,
                f"{http_resp.status}: {http_resp.content.decode('utf-8')}",
            )
    else:
        return make_error(error)


def make_error(error):
    status = None
    try:
        status = Code(error["code"])
    except KeyError:
        status = Code.unknown
        pass

    return ConnectException(status, error.get("message", ""))


class GzipCompressor:
    name = "gzip"
    decompress = gzip.decompress
    compress = gzip.compress


class JSONCodec:
    content_type = "json"

    @staticmethod
    def encode(msg):
        return json_format.MessageToJson(msg).encode("utf8")

    @staticmethod
    def decode(data, *, msg_type):
        msg = msg_type()
        json_format.Parse(data.decode("utf8"), msg)
        return msg


class ProtobufCodec:
    content_type = "proto"

    @staticmethod
    def encode(msg):
        return msg.SerializeToString()

    @staticmethod
    def decode(data, *, msg_type):
        msg = msg_type()
        msg.ParseFromString(data)
        return msg


class Client:
    def __init__(
        self,
        *,
        pool: Optional[ConnectionPool] = None,
        async_pool: Optional[AsyncConnectionPool] = None,
        url: str,
        response_type,
        compressor=None,
        json: Optional[bool] = False,
        headers: Optional[Dict[str, str]] = None,
    ):
        if headers is None:
            headers = {}

        self.pool = pool
        self.async_pool = async_pool
        self.url = url
        self._codec = JSONCodec if json else ProtobufCodec
        self._response_type = response_type
        self._compressor = compressor
        self._headers = {**{"user-agent": "connect-python"}, **headers}

    def _prepare_unary_request(
        self,
        req,
        request_timeout=None,
        headers={},
        **opts,
    ):
        data = self._codec.encode(req)

        if self._compressor is not None:
            data = self._compressor.compress(data)

        extensions = (
            None
            if request_timeout is None
            else {
                "timeout": {
                    "connect": request_timeout,
                    "pool": request_timeout,
                    "read": request_timeout,
                    "write": request_timeout,
                }
            }
        )

        return {
            "method": "POST",
            "url": self.url,
            "content": data,
            "extensions": extensions,
            "headers": {
                **self._headers,
                **headers,
                **opts.get("headers", {}),
                "connect-protocol-version": "1",
                "content-encoding": (
                    "identity" if self._compressor is None else self._compressor.name
                ),
                "content-type": f"application/{self._codec.content_type}",
            },
        }

    def _process_unary_response(
        self,
        http_resp: Response,
    ):
        if http_resp.status != 200:
            raise error_for_response(http_resp)

        content = http_resp.content

        if self._compressor is not None:
            content = self._compressor.decompress(content)

        return self._codec.decode(
            content,
            msg_type=self._response_type,
        )

    async def acall_unary(
        self,
        req,
        request_timeout=None,
        headers={},
        **opts,
    ):
        if self.async_pool is None:
            raise ValueError("async_pool is required")

        res = await self.async_pool.request(
            **self._prepare_unary_request(
                req,
                request_timeout,
                headers,
                **opts,
            )
        )

        return self._process_unary_response(res)

    def call_unary(self, req, request_timeout=None, headers={}, **opts):
        if self.pool is None:
            raise ValueError("pool is required")

        res = self.pool.request(
            **self._prepare_unary_request(
                req,
                request_timeout,
                headers,
                **opts,
            )
        )

        return self._process_unary_response(res)

    def _create_stream_timeout(self, timeout: Optional[int]):
        if timeout:
            return {"connect-timeout-ms": str(timeout * 1000)}
        return {}

    def _prepare_server_stream_request(
        self,
        req,
        request_timeout=None,
        timeout=None,
        headers={},
        **opts,
    ):
        data = self._codec.encode(req)
        flags = EnvelopeFlags(0)

        extensions = (
            None
            if request_timeout is None
            else {"timeout": {"connect": request_timeout, "pool": request_timeout}}
        )

        if self._compressor is not None:
            data = self._compressor.compress(data)
            flags |= EnvelopeFlags.compressed

        stream_timeout = self._create_stream_timeout(timeout)

        return {
            "method": "POST",
            "url": self.url,
            "content": encode_envelope(
                flags=flags,
                data=data,
            ),
            "extensions": extensions,
            "headers": {
                **self._headers,
                **headers,
                **opts.get("headers", {}),
                **stream_timeout,
                "connect-protocol-version": "1",
                "connect-content-encoding": (
                    "identity" if self._compressor is None else self._compressor.name
                ),
                "content-type": f"application/connect+{self._codec.content_type}",
            },
        }

    async def acall_server_stream(
        self,
        req,
        request_timeout=None,
        timeout=None,
        headers={},
        **opts,
    ):
        if self.async_pool is None:
            raise ValueError("async_pool is required")

        async with self.async_pool.stream(
            **self._prepare_server_stream_request(
                req,
                request_timeout,
                timeout,
                headers,
                **opts,
            )
        ) as http_resp:
            if http_resp.status != 200:
                raise error_for_response(http_resp)

            buffer = b""
            end_stream = False
            needs_header = True
            flags, data_len = 0, 0

            async for chunk in http_resp.aiter_stream():
                buffer += chunk

                if needs_header:
                    header = buffer[:envelope_header_length]
                    buffer = buffer[envelope_header_length:]
                    flags, data_len = decode_envelope_header(header)
                    needs_header = False
                    end_stream = EnvelopeFlags.end_stream in flags

                if len(buffer) >= data_len:
                    buffer = buffer[:data_len]

                    if end_stream:
                        data = (
                            buffer
                            if self._compressor is None
                            else self._compressor.decompress(buffer)
                        )

                        data = json.loads(data)

                        if "error" in data:
                            raise make_error(data["error"])

                        # TODO: Figure out what else might be possible
                        return

                    if self._compressor is not None:
                        buffer = self._compressor.decompress(buffer)

                    # TODO: handle server message compression
                    # if EnvelopeFlags.compression in flags:
                    # TODO: should the client potentially use a different codec
                    # based on response header? Or can we assume they're always
                    # the same and an error otherwise.
                    yield self._codec.decode(buffer, msg_type=self._response_type)

                    buffer = buffer[data_len:]
                    needs_header = True

    def call_server_stream(
        self,
        req,
        request_timeout=None,
        timeout=None,
        headers={},
        **opts,
    ):
        if self.pool is None:
            raise ValueError("pool is required")

        with self.pool.stream(
            **self._prepare_server_stream_request(
                req,
                request_timeout,
                timeout,
                headers,
                **opts,
            )
        ) as http_resp:
            if http_resp.status != 200:
                raise error_for_response(http_resp)

            buffer = b""
            end_stream = False
            needs_header = True
            flags, data_len = 0, 0

            for chunk in http_resp.iter_stream():
                buffer += chunk

                if needs_header:
                    header = buffer[:envelope_header_length]
                    buffer = buffer[envelope_header_length:]
                    flags, data_len = decode_envelope_header(header)
                    needs_header = False
                    end_stream = EnvelopeFlags.end_stream in flags

                if len(buffer) >= data_len:
                    buffer = buffer[:data_len]

                    if end_stream:
                        data = (
                            buffer
                            if self._compressor is None
                            else self._compressor.decompress(buffer)
                        )

                        data = json.loads(data)

                        if "error" in data:
                            raise make_error(data["error"])

                        # TODO: Figure out what else might be possible
                        return

                    if self._compressor is not None:
                        buffer = self._compressor.decompress(buffer)

                    # TODO: handle server message compression
                    # if EnvelopeFlags.compression in flags:
                    # TODO: should the client potentially use a different codec
                    # based on response header? Or can we assume they're always
                    # the same and an error otherwise.
                    yield self._codec.decode(buffer, msg_type=self._response_type)

                    buffer = buffer[data_len:]
                    needs_header = True

    def call_client_stream(self, req, **opts):
        raise NotImplementedError("client stream not supported")

    def acall_client_stream(self, req, **opts):
        raise NotImplementedError("client stream not supported")

    def call_bidi_stream(self, req, **opts):
        raise NotImplementedError("bidi stream not supported")
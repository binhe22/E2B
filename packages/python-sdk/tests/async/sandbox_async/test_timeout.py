import pytest

from time import sleep

from e2b.exceptions import TimeoutException
from e2b import AsyncSandbox


@pytest.mark.skip_debug()
@pytest.mark.asyncio
async def test_shorten_timeout(async_sandbox: AsyncSandbox):
    await async_sandbox.set_timeout(5)
    sleep(6)
    with pytest.raises(TimeoutException):
        await async_sandbox.is_running()


@pytest.mark.skip_debug()
@pytest.mark.asyncio
async def test_shorten_then_lengthen_timeout(async_sandbox: AsyncSandbox):
    await async_sandbox.set_timeout(5)
    sleep(1)
    await async_sandbox.set_timeout(10)
    sleep(6)
    await async_sandbox.is_running()
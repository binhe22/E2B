import {
  useState,
  useEffect,
  Fragment,
} from 'react'
import { useRouter } from 'next/router'
import clsx from 'clsx'
import {
  SystemContext,
  UserContext,
  AssistantContext,
} from 'utils/agentLogs'

export interface Props {
  context: (SystemContext | UserContext | AssistantContext)[]
  onSelected: (ctx: SystemContext | UserContext | AssistantContext) => void
}

function AgentContext({
  context,
  onSelected,
}: Props) {
  const router = useRouter()
  const [opened, setOpened] = useState<number>()

  useEffect(function selectLogBasedOnURLQuery() {
    const selectedLog = router.query.selectedLog as string
    let idx: number
    if (selectedLog) {
      idx = parseInt(selectedLog)
    } else {
      idx = 0
    }
    setOpened(idx)
    onSelected(context[idx])
  }, [router, context, onSelected])

  function open(idx: number) {
    setOpened(idx)
    onSelected(context[idx])
    router.push(`/log/${router.query.logFileID}?selectedLog=${idx}`, undefined, { shallow: true })
  }

  function close() {
    setOpened(undefined)
  }

  function toggle(idx: number) {
    if (opened === idx) {
      close()
    } else {
      open(idx)
    }
  }

  return (
    <div className="flex-1 flex flex-col space-y-2 max-w-full w-full overflow-hidden">
      <h2 className="font-medium text-sm text-gray-500">Logs</h2>

      <div className="flex-1 flex flex-col space-y-1 max-w-full w-full overflow-auto">
        {context.map((ctx, idx) => (
          <Fragment key={idx}>
            <div className="flex items-center space-x-2 ">
              <span className={clsx(
                'font-bold text-sm capitalize min-w-[72px]',
                opened === idx && 'text-[#6366F1]',
                opened !== idx && 'text-[#55618C]',
              )}
              >
                {ctx.role}
              </span>
              <span
                className={clsx(
                  'text-sm text-gray-100 max-w-full truncate p-2 hover:bg-[#1F2437] transition-all rounded-md cursor-pointer w-full',
                  opened === idx && 'bg-[#1F2437]',
                )}
                onClick={() => toggle(idx)}
              >
                {ctx.content}
              </span>
            </div>
            {idx !== context.length - 1 && (
              <div className="ml-1 rounded min-h-[20px] w-px bg-gray-800" />
            )}
          </Fragment>
        ))}
      </div>
    </div>
  )
}

export default AgentContext
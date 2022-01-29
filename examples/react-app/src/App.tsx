import {
  useState,
  useCallback,
  useEffect,
} from 'react';

import {
  useDevbook,
  Env,
  DevbookStatus,
} from '@devbookhq/sdk';
import Splitter from '@devbookhq/splitter';

import './App.css';
import Editor from './Editor';
import Output from './Output';

const initialCode = `const os = require('os');
console.log('Hostname:', os.hostname());
console.log(process.env)`

const initialCmd =
  `ls -l
`

function App() {
  const [sizes, setSizes] = useState([50, 50]);
  const [code, setCode] = useState(initialCode);
  const [cmd, setCmd] = useState(initialCmd);
  const [execType, setExecType] = useState('code');

  const {
    stderr,
    stdout,
    runCode,
    runCmd,
    status,
    fs,
  } = useDevbook({ debug: true, env: Env.Supabase });
  useDevbook({ debug: true, env: Env.Supabase });
  useDevbook({ debug: true, env: Env.Supabase });
  console.log({ stdout, stderr });

  async function getFile() {
    if (!fs) return
    if (status !== DevbookStatus.Connected) return

    setInterval(async () => {
      const random = Math.random()
      await fs.write('/src/package.json', random.toString())
      const content = await fs.get('/src/package.json', 'content')
      console.log('content', content)
    }, 2000)
  }

  const handleEditorChange = useCallback((content: string) => {
    if (execType === 'code') {
      setCode(content);
    } else {
      setCmd(content);
    }
  }, [setCode, execType]);

  const run = useCallback(() => {
    if (execType === 'code') {
      runCode(code);
    } else {
      runCmd(cmd);
    }
  }, [runCode, runCmd, code, cmd, execType]);


  return (
    <div className="app">
      {status === DevbookStatus.Disconnected && <div>Status: Disconnected, will start VM</div>}
      {status === DevbookStatus.Connecting && <div>Status: Starting VM...</div>}
      {status === DevbookStatus.Connected && (
        <div className="controls">
          <select className="type" value={execType} onChange={e => setExecType(e.target.value)}>
            <option value="code">Code</option>
            <option value="cmd">Command</option>
          </select>
          <button className="run-btn" onClick={getFile}>Run</button>
        </div>
      )}

      <Splitter
        classes={['flex', 'flex']}
        initialSizes={sizes}
        onResizeFinished={(_, sizes) => {
          setSizes(sizes);
        }}
      >
        <Editor
          initialCode={execType === 'code' ? initialCode : initialCmd}
          onChange={handleEditorChange}
        />
        <Output
          stdout={stdout}
          stderr={stderr}
        />
      </Splitter>
    </div >
  );
}

export default App;

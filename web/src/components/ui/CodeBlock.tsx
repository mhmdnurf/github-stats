import { useState } from 'react'

type CodeBlockProps = {
  code: string
  language?: string
}

export function CodeBlock({ code, language = 'shell' }: CodeBlockProps) {
  const [copied, setCopied] = useState(false)

  async function handleCopy() {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className="code-block">
      <div className="code-block-head">
        <span className="code-block-lang">{language}</span>
        <button type="button" className="code-block-copy" onClick={handleCopy}>
          {copied ? 'Copied' : 'Copy'}
        </button>
      </div>
      <pre>
        <code>{code}</code>
      </pre>
    </div>
  )
}

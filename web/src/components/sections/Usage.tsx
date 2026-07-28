import { Section } from '../ui/Section'
import { CodeBlock } from '../ui/CodeBlock'
import { Table } from '../ui/Table'

const PARAM_ROWS = [
  { key: '1', param: <code>theme</code>, required: 'No', default: <code>default</code>, description: 'SVG theme for every card endpoint' },
  { key: '2', param: <code>repositories</code>, required: 'No', default: <code>public</code>, description: 'public or all; dynamic username endpoints only accept public' },
]

export function Usage() {
  return (
    <Section
      id="usage"
      eyebrow="Usage"
      title="Embed cards in Markdown"
      description="Use the short endpoints for the configured account, or include a GitHub username in the path to render a public card for any account."
    >
      <div className="step-list">
        <div className="step">
          <h3>Statistics card</h3>
          <CodeBlock code="https://your-domain/stats" language="text" />
        </div>
        <div className="step">
          <h3>Languages card</h3>
          <CodeBlock code="https://your-domain/languages" language="text" />
        </div>
        <div className="step">
          <h3>Cards for any GitHub username</h3>
          <CodeBlock
            code={`https://your-domain/your-username/stats\nhttps://your-domain/your-username/languages?theme=light`}
            language="text"
          />
          <p className="step-note">
            These endpoints fetch public GitHub data on demand, cache it in memory for five minutes,
            and are rate-limited per client IP. They reject <code>repositories=all</code>.
          </p>
        </div>
        <div className="step">
          <h3>Markdown embed</h3>
          <CodeBlock
            language="markdown"
            code={`![GitHub statistics](https://your-domain/your-username/stats)\n![Most used languages](https://your-domain/your-username/languages)`}
          />
        </div>
        <div className="step">
          <h3>Query parameters</h3>
          <Table
            columns={[
              { key: 'param', header: 'Parameter' },
              { key: 'required', header: 'Required' },
              { key: 'default', header: 'Default' },
              { key: 'description', header: 'Description' },
            ]}
            rows={PARAM_ROWS}
          />
        </div>
      </div>
    </Section>
  )
}

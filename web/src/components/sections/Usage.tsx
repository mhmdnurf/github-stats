import { Section } from '../ui/Section'
import { CodeBlock } from '../ui/CodeBlock'
import { Table } from '../ui/Table'

const PARAM_ROWS = [
  { key: '1', param: <code>theme</code>, required: 'No', default: <code>default</code>, description: 'SVG card theme for /stats and /languages' },
  { key: '2', param: <code>repositories</code>, required: 'No', default: <code>public</code>, description: 'Repository scope: public or all' },
]

export function Usage() {
  return (
    <Section
      id="usage"
      eyebrow="Usage"
      title="Embed cards in Markdown"
      description="The GitHub username is configured through GITHUB_USERNAME and cannot be overridden through query parameters."
    >
      <div className="step-list">
        <div className="step">
          <h3>Statistics card</h3>
          <CodeBlock code="https://your-domain.example/stats" language="text" />
        </div>
        <div className="step">
          <h3>Languages card</h3>
          <CodeBlock code="https://your-domain.example/languages" language="text" />
        </div>
        <div className="step">
          <h3>Markdown embed</h3>
          <CodeBlock
            language="markdown"
            code={`![GitHub statistics](https://your-domain.example/stats)\n![Most used languages](https://your-domain.example/languages)`}
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

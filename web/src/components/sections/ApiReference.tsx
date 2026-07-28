import { Section } from '../ui/Section'
import { CodeBlock } from '../ui/CodeBlock'
import { Table } from '../ui/Table'

const ERROR_ROWS = [
  { key: '400', status: <code>400</code>, meaning: 'Unknown theme or invalid repositories value' },
  { key: '503', status: <code>503</code>, meaning: 'A snapshot is not available yet' },
  { key: '504', status: <code>504</code>, meaning: 'Snapshot storage exceeded the deadline' },
  { key: '500', status: <code>500</code>, meaning: 'Unexpected server error' },
]

export function ApiReference() {
  return (
    <Section id="api" eyebrow="API reference" title="Endpoints">
      <div className="step-list">
        <div className="step">
          <h3>Generate a statistics card</h3>
          <CodeBlock language="http" code="GET /stats?theme={theme}&repositories={public|all}" />
        </div>
        <div className="step">
          <h3>Generate a languages card</h3>
          <CodeBlock language="http" code="GET /languages?theme={theme}&repositories={public|all}" />
        </div>
        <div className="step">
          <h3>Health check</h3>
          <CodeBlock language="http" code="GET /health" />
          <p className="step-note">Returns 200 OK when the server is running.</p>
        </div>
        <div className="step">
          <h3>Error responses</h3>
          <Table
            columns={[
              { key: 'status', header: 'Status' },
              { key: 'meaning', header: 'Meaning' },
            ]}
            rows={ERROR_ROWS}
          />
        </div>
      </div>
    </Section>
  )
}

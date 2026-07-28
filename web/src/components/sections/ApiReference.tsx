import { Section } from '../ui/Section'
import { CodeBlock } from '../ui/CodeBlock'
import { Table } from '../ui/Table'

const ERROR_ROWS = [
  { key: '400', status: <code>400</code>, meaning: 'Unknown theme, invalid username or repositories value, or repositories=all on a dynamic endpoint' },
  { key: '404', status: <code>404</code>, meaning: 'GitHub user not found on a dynamic endpoint' },
  { key: '429', status: <code>429</code>, meaning: 'Per-IP dynamic endpoint rate limit exceeded; retry after one second' },
  { key: '503', status: <code>503</code>, meaning: 'A configured-account snapshot is not available yet' },
  { key: '504', status: <code>504</code>, meaning: 'Snapshot storage or a live GitHub request exceeded the deadline' },
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
          <h3>Generate cards for any public GitHub account</h3>
          <CodeBlock
            language="http"
            code={`GET /{username}/stats?theme={theme}&repositories=public\nGET /{username}/languages?theme={theme}&repositories=public`}
          />
          <p className="step-note">
            Dynamic endpoints fetch live GitHub data, accept only public repository scope,
            and share a per-IP limiter with a burst of 10 requests that refills at one request per second.
          </p>
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

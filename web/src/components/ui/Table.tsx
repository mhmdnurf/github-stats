import type { ReactNode } from 'react'

export type TableColumn = {
  key: string
  header: string
}

type TableProps = {
  columns: TableColumn[]
  rows: Record<string, ReactNode>[]
}

export function Table({ columns, rows }: TableProps) {
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            {columns.map((col) => (
              <th key={col.key}>{col.header}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i}>
              {columns.map((col) => (
                <td key={col.key}>{row[col.key]}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

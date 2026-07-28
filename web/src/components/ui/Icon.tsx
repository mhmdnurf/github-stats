type IconProps = {
  name: string
  className?: string
}

export function Icon({ name, className }: IconProps) {
  return (
    <svg className={['icon', className].filter(Boolean).join(' ')} role="presentation" aria-hidden="true">
      <use href={`/icons.svg#${name}`} />
    </svg>
  )
}

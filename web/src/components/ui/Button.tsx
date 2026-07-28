import type { AnchorHTMLAttributes, ButtonHTMLAttributes, ReactNode } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost'

const baseClass = 'btn'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: Variant
  icon?: ReactNode
}

export function Button({ variant = 'secondary', icon, className, children, ...rest }: ButtonProps) {
  return (
    <button className={[baseClass, `btn-${variant}`, className].filter(Boolean).join(' ')} {...rest}>
      {icon}
      {children}
    </button>
  )
}

type LinkButtonProps = AnchorHTMLAttributes<HTMLAnchorElement> & {
  variant?: Variant
  icon?: ReactNode
}

export function LinkButton({ variant = 'secondary', icon, className, children, ...rest }: LinkButtonProps) {
  return (
    <a className={[baseClass, `btn-${variant}`, className].filter(Boolean).join(' ')} {...rest}>
      {icon}
      {children}
    </a>
  )
}

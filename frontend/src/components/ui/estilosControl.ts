export const estiloControl =
  'h-12 w-full rounded-xl border bg-surface-container-lowest px-4 text-body-md text-on-surface shadow-sm ' +
  'placeholder:text-outline focus:outline-none focus:ring-2 focus:ring-primary-container ' +
  'disabled:cursor-not-allowed disabled:bg-surface-container-low disabled:text-on-surface-variant'

export function bordeControl(invalido?: boolean) {
  return invalido ? 'border-error' : 'border-outline-variant'
}

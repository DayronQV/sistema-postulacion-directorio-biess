import Input, { type InputProps } from './Input'

// Campo de fecha nativo (el valor viaja como AAAA-MM-DD, el navegador lo muestra según su idioma).
export default function DateInput(props: Omit<InputProps, 'type'>) {
  return <Input type="date" {...props} />
}

export default function PieDePagina() {
  return (
    <footer className="bg-surface-container-lowest px-margin-mobile py-space-lg shadow-[0_-1px_8px_rgba(0,0,0,0.04)] sm:px-margin">
      <p className="text-center text-legal-code text-on-surface-variant">
        © {new Date().getFullYear()} Banco del Instituto Ecuatoriano de Seguridad Social (BIESS).
        Sistema del Concurso de Méritos y Oposición.
      </p>
    </footer>
  )
}

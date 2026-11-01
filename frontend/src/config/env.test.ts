import { env } from '@/config/env'

describe('env', () => {
  it('expone la URL base de la API', () => {
    expect(env.apiUrl).toMatch(/^https?:\/\//)
  })
})

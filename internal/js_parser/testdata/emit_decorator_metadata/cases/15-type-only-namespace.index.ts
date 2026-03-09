import type { Services } from './15-type-only-namespace.services'

declare const decorator: any

class Main {
  @decorator
  field: Services.Service
}

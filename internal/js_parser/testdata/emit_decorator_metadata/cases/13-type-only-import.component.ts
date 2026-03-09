import type { Service } from './13-type-only-import.service'

declare const decorator: any

@decorator
class MyComponent {
  constructor(public service: Service) {}

  @decorator
  method(x: this) {}
}

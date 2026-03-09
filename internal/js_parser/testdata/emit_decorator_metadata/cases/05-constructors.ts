function dec(...args: any[]) {}

@dec
class ExplicitEmpty {
  constructor() {}
}

@dec
class ImplicitConstructor {}

class ParameterDecorated {
  constructor(@dec a: string, @dec b: number) {}
}

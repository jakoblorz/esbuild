function dec(...args: any[]) {}

class Accessors {
  @dec
  get typedGet(): string {
    return ''
  }

  @dec
  get untypedGet() {
    return ''
  }

  @dec
  set typedSet(value: string) {}

  @dec
  set untypedSet(value) {}
}

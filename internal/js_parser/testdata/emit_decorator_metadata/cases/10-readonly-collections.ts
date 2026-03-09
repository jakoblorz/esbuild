function dec(...args: any[]) {}

class ReadonlyCollections {
  @dec roArray: readonly string[]
  @dec roTuple: readonly [string, number]
  @dec roArrayRef: ReadonlyArray<string>
}

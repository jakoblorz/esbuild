function dec(...args: any[]) {}

class AsyncMethods {
  @dec
  async inferred() {}

  @dec
  async explicit(): Promise<number> {
    return 1
  }

  @dec
  syncPromise(input: Promise<number>): Promise<number> {
    return input
  }
}

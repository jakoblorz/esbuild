function dec(...args) {
}
class AsyncMethods {
  async inferred() {
  }
  async explicit() {
    return 1;
  }
  syncPromise(input) {
    return input;
  }
}
__decorateClass([
  dec,
  __legacyMetadata("design:type", Function),
  __legacyMetadata("design:paramtypes", []),
  __legacyMetadata("design:returntype", Promise)
], AsyncMethods.prototype, "inferred", 1);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Function),
  __legacyMetadata("design:paramtypes", []),
  __legacyMetadata("design:returntype", Promise)
], AsyncMethods.prototype, "explicit", 1);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Function),
  __legacyMetadata("design:paramtypes", [
    Promise
  ]),
  __legacyMetadata("design:returntype", Promise)
], AsyncMethods.prototype, "syncPromise", 1);

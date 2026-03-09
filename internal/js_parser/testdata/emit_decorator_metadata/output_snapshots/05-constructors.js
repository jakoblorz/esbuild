function dec(...args) {
}
let ExplicitEmpty = class {
  constructor() {
  }
};
ExplicitEmpty = __decorateClass([
  dec,
  __legacyMetadata("design:paramtypes", [])
], ExplicitEmpty);
let ImplicitConstructor = class {
};
ImplicitConstructor = __decorateClass([
  dec
], ImplicitConstructor);
let ParameterDecorated = class {
  constructor(a, b) {
  }
};
ParameterDecorated = __decorateClass([
  __decorateParam(0, dec),
  __decorateParam(1, dec),
  __legacyMetadata("design:paramtypes", [
    String,
    Number
  ])
], ParameterDecorated);

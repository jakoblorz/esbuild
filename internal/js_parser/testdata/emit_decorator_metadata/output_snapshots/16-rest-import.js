// ----- 16-rest-import.aux.ts -----
export class SomeClass {
}

// ----- 16-rest-import.main.ts -----
let ClassA = class {
  constructor(...init) {
  }
  foo(...args) {
  }
};
__decorateClass([
  annotation1,
  __legacyMetadata("design:type", Function),
  __legacyMetadata("design:paramtypes", [
    SomeClass
  ]),
  __legacyMetadata("design:returntype", void 0)
], ClassA.prototype, "foo", 1);
ClassA = __decorateClass([
  annotation,
  __legacyMetadata("design:paramtypes", [
    SomeClass
  ])
], ClassA);

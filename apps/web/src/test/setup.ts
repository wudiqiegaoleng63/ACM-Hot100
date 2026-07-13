import '@testing-library/jest-dom/vitest';

Object.defineProperties(window, {
  AbortController: { configurable: true, value: globalThis.AbortController },
  AbortSignal: { configurable: true, value: globalThis.AbortSignal },
});

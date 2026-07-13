import '@testing-library/jest-dom/vitest';

Object.defineProperties(window, {
  AbortController: { configurable: true, value: globalThis.AbortController },
  AbortSignal: { configurable: true, value: globalThis.AbortSignal },
  matchMedia: {
    configurable: true,
    value: (query: string): MediaQueryList => ({
      matches: false,
      media: query,
      onchange: null,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      addListener: () => undefined,
      removeListener: () => undefined,
      dispatchEvent: () => false,
    }),
  },
});

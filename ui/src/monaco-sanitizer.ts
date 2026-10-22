import createDOMPurify from 'dompurify';

// Monaco's bundled sanitizer is outdated. Keep its hooks separate from other renderers.
const purifier = createDOMPurify(window);

/** Sanitize markup rendered by Monaco. */
export const sanitize = purifier.sanitize;
/** Register Monaco's attribute and URL policy hooks. */
export const addHook = purifier.addHook;
/** Remove Monaco's temporary policy hook after rendering. */
export const removeHook = purifier.removeHook;

/** Isolated sanitizer instance used by current Monaco releases. */
export default purifier;

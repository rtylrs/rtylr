import { describe, expect, it } from 'vitest';

import { isPagerPublicPath } from './pagerPaths';

describe('isPagerPublicPath', () => {
  it.each(['/acme/pager', '/acme/pager/', '/pager/t/abc123', '/pager/pager'])(
    'treats %s as a customer page',
    (path) => {
      expect(isPagerPublicPath(path)).toBe(true);
    }
  );

  it.each(['/', '/pager', '/pager/', '/pager/poster', '/acme/menu', '/acme/pager/extra'])(
    'leaves %s alone',
    (path) => {
      expect(isPagerPublicPath(path)).toBe(false);
    }
  );
});

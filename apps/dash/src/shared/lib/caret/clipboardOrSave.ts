import { downloadTextFile } from './download';
import { safeFilename } from './safeFilename';

/** Copy `text` to the clipboard, falling back to a download when the
 *  Clipboard API is unavailable. */
export async function copyOrDownload(
  name: string,
  text: string,
  mime = 'text/plain'
): Promise<'clipboard' | 'download'> {
  try {
    await navigator.clipboard.writeText(text);
    return 'clipboard';
  } catch {
    downloadTextFile(safeFilename(name, mime === 'text/markdown' ? 'md' : 'txt'), text, mime);
    return 'download';
  }
}

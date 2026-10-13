/** Text files Caret reads: exports, notes, data and code. */
const TEXT_EXTENSIONS = [
  "csv", "tsv", "txt", "md", "markdown", "json", "xml", "html", "htm", "log",
  "yaml", "yml", "ini", "sql", "js", "ts", "tsx", "jsx", "py", "go", "java",
  "rb", "php", "css", "sh", "ics", "vcf", "srt",
];

/** Binary documents the server turns into text: PDF, Word, Excel, PowerPoint. */
const DOCUMENT_EXTENSIONS = ["pdf", "docx", "xlsx", "pptx", "odt", "ods", "odp"];

export const TEXT_FILE_ACCEPT = [...TEXT_EXTENSIONS, ...DOCUMENT_EXTENSIONS].map((ext) => `.${ext}`).join(",");
export const IMAGE_ACCEPT = "image/jpeg,image/png,image/webp";
export const MAX_TEXT_FILES = 3;
const MAX_TEXT_FILE_BYTES = 400 * 1024;
const MAX_DOCUMENT_BYTES = 2 * 1024 * 1024;
/** The API accepts 4 MiB per request; leave room for the prompt. */
export const MAX_ATTACHMENT_PAYLOAD_BYTES = 3.6 * 1024 * 1024;

export interface CaretTextFile {
  name: string;
  text: string;
  /** A document (PDF, Word, Excel, PowerPoint) is sent as data; the server extracts its text. */
  data_url?: string;
}

function extensionOf(file: File): string {
  return file.name.split(".").pop()?.toLowerCase() ?? "";
}

function isPDF(file: File): boolean {
  return file.type === "application/pdf" || extensionOf(file) === "pdf";
}

function isDocument(file: File): boolean {
  return isPDF(file) || DOCUMENT_EXTENSIONS.includes(extensionOf(file));
}

function readDataUrl(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result));
    reader.onerror = () => reject(new Error(`${file.name} could not be read.`));
    reader.readAsDataURL(file);
  });
}

export function isImageFile(file: File): boolean {
  return ["image/jpeg", "image/png", "image/webp"].includes(file.type);
}

export function isTextFile(file: File): boolean {
  if (
    file.type.startsWith("text/") ||
    file.type === "application/json" ||
    file.type === "application/pdf"
  )
    return true;
  const ext = extensionOf(file);
  return TEXT_EXTENSIONS.includes(ext) || DOCUMENT_EXTENSIONS.includes(ext);
}

/** Reads attached text files, enforcing the server's limits up front. */
export async function readTextFiles(files: File[]): Promise<CaretTextFile[]> {
  if (files.length > MAX_TEXT_FILES)
    throw new Error(`Attach up to ${MAX_TEXT_FILES} files per message.`);
  return Promise.all(
    files.map(async (file) => {
      if (isDocument(file)) {
        if (file.size > MAX_DOCUMENT_BYTES)
          throw new Error(`${file.name} is larger than 2 MB.`);
        const dataUrl = await readDataUrl(file);
        return {
          name: file.name,
          text: "",
          // The server reads a PDF by its media type, an Office file by its name.
          data_url: isPDF(file)
            ? dataUrl.replace(/^data:[^;]*;/, "data:application/pdf;")
            : dataUrl,
        };
      }
      if (file.size > MAX_TEXT_FILE_BYTES)
        throw new Error(
          `${file.name} is larger than 400 KB; attach a smaller extract.`,
        );
      return { name: file.name, text: await file.text() };
    }),
  );
}

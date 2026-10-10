/** Text files Caret reads: exports, notes, data and code. */
const TEXT_EXTENSIONS = [
  "csv", "tsv", "txt", "md", "markdown", "json", "xml", "html", "htm", "log",
  "yaml", "yml", "ini", "sql", "js", "ts", "tsx", "jsx", "py", "go", "java",
  "rb", "php", "css", "sh", "ics", "vcf", "srt",
];

export const TEXT_FILE_ACCEPT = TEXT_EXTENSIONS.map((ext) => `.${ext}`).join(",");
export const IMAGE_ACCEPT = "image/jpeg,image/png,image/webp";
export const MAX_TEXT_FILES = 3;
const MAX_TEXT_FILE_BYTES = 400 * 1024;

export interface CaretTextFile {
  name: string;
  text: string;
}

export function isImageFile(file: File): boolean {
  return ["image/jpeg", "image/png", "image/webp"].includes(file.type);
}

export function isTextFile(file: File): boolean {
  if (file.type.startsWith("text/") || file.type === "application/json")
    return true;
  const ext = file.name.split(".").pop()?.toLowerCase() ?? "";
  return TEXT_EXTENSIONS.includes(ext);
}

/** Reads attached text files, enforcing the server's limits up front. */
export async function readTextFiles(files: File[]): Promise<CaretTextFile[]> {
  if (files.length > MAX_TEXT_FILES)
    throw new Error(`Attach up to ${MAX_TEXT_FILES} files per message.`);
  return Promise.all(
    files.map(async (file) => {
      if (file.size > MAX_TEXT_FILE_BYTES)
        throw new Error(
          `${file.name} is larger than 400 KB; attach a smaller extract.`,
        );
      return { name: file.name, text: await file.text() };
    }),
  );
}

import type { CaretImageAttachment } from "../types/message";

const MAX_IMAGE_BYTES = 2 * 1024 * 1024;
const MAX_IMAGE_COUNT = 4;
const MAX_EDGE = 1600;

export async function prepareCaretImages(
  files: File[],
): Promise<CaretImageAttachment[]> {
  if (files.length > MAX_IMAGE_COUNT)
    throw new Error("Attach up to four images per message.");
  const images = await Promise.all(files.map(prepareCaretImage));
  const totalBytes = images.reduce(
    (total, image) => total + estimateDataUrlBytes(image.dataUrl),
    0,
  );
  if (totalBytes > MAX_IMAGE_BYTES)
    throw new Error("The resized images exceed the 2 MB attachment limit.");
  return images;
}

async function prepareCaretImage(file: File): Promise<CaretImageAttachment> {
  if (!["image/jpeg", "image/png", "image/webp"].includes(file.type)) {
    throw new Error("Caret accepts JPEG, PNG, and WebP images.");
  }
  if (file.size > 12 * 1024 * 1024)
    throw new Error(`${file.name} is too large to process.`);
  if (typeof createImageBitmap !== "function")
    throw new Error("This browser cannot resize image attachments.");

  const bitmap = await createImageBitmap(file);
  try {
    if (bitmap.width * bitmap.height > 40_000_000)
      throw new Error(`${file.name} has too many pixels.`);
    const scale = Math.min(1, MAX_EDGE / Math.max(bitmap.width, bitmap.height));
    const canvas = document.createElement("canvas");
    canvas.width = Math.max(1, Math.round(bitmap.width * scale));
    canvas.height = Math.max(1, Math.round(bitmap.height * scale));
    const context = canvas.getContext("2d");
    if (!context) throw new Error("Could not resize the image.");
    context.drawImage(bitmap, 0, 0, canvas.width, canvas.height);

    const outputType = file.type === "image/png" ? "image/png" : "image/jpeg";
    const initialQuality = outputType === "image/jpeg" ? 0.84 : undefined;
    let blob = await canvasBlob(canvas, outputType, initialQuality);
    for (const quality of [0.7, 0.55, 0.4]) {
      if (
        blob.size <= MAX_IMAGE_BYTES / MAX_IMAGE_COUNT ||
        outputType !== "image/jpeg"
      )
        break;
      blob = await canvasBlob(canvas, outputType, quality);
    }
    if (blob.size > MAX_IMAGE_BYTES)
      throw new Error(`${file.name} is still too large after resizing.`);
    return { name: file.name.slice(0, 200), dataUrl: await blobDataUrl(blob) };
  } finally {
    bitmap.close();
  }
}

function canvasBlob(
  canvas: HTMLCanvasElement,
  type: string,
  quality?: number,
): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) =>
        blob ? resolve(blob) : reject(new Error("Could not encode the image.")),
      type,
      quality,
    );
  });
}

function blobDataUrl(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () =>
      reject(new Error("Could not read the image attachment."));
    reader.onload = () =>
      typeof reader.result === "string"
        ? resolve(reader.result)
        : reject(new Error("Invalid image data."));
    reader.readAsDataURL(blob);
  });
}

function estimateDataUrlBytes(dataUrl: string): number {
  const encoded = dataUrl.slice(dataUrl.indexOf(",") + 1);
  return Math.floor((encoded.length * 3) / 4);
}

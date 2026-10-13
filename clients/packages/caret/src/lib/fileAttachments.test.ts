import { describe, expect, it } from "vitest";
import { isTextFile, readTextFiles } from "./fileAttachments";

const DOCX_TYPE =
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document";

describe("document attachments", () => {
  it("accepts Word, Excel and PowerPoint files by extension", () => {
    for (const name of ["memo.docx", "budget.XLSX", "deck.pptx", "notes.odt"]) {
      expect(isTextFile(new File(["x"], name))).toBe(true);
    }
    expect(isTextFile(new File(["x"], "legacy.doc"))).toBe(false);
  });

  it("sends Office files as data so the server can extract their text", async () => {
    const [docx] = await readTextFiles([
      new File(["PK"], "memo.docx", { type: DOCX_TYPE }),
    ]);
    expect(docx.text).toBe("");
    expect(docx.data_url).toMatch(/^data:application\/vnd\.openxmlformats.*;base64,/);

    const [pdf] = await readTextFiles([new File(["%PDF"], "a.pdf")]);
    expect(pdf.data_url).toMatch(/^data:application\/pdf;base64,/);
  });

  it("rejects documents over 2 MB before sending", async () => {
    const big = new File([new Uint8Array(2 * 1024 * 1024 + 1)], "big.xlsx");
    await expect(readTextFiles([big])).rejects.toThrow("larger than 2 MB");
  });
});

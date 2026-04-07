import React from "react";
import { useState, useRef, useEffect } from "react";
import { QRCode } from "react-qrcode-logo";
import html2canvas from "html2canvas";
import { saveAs } from "file-saver";
import { useUserStore } from "../../../stores/userStore";
import { QrCodeTemplate, QrCodeTemplates, QrOptions } from "./templates";
import styles from "./QrCode.module.scss";

function QrCode() {
  const { userId, user } = useUserStore();
  const [qrOptions, setQrOptions] = useState<QrOptions>({
    url: "https://rtylr.voxire.com/restaurant/" + userId,
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffffff",
    logoImage: false,
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: "squares",
  });
  const [isLogoLoaded, setIsLogoLoaded] = useState(false);

  const qrRef = useRef(null);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setQrOptions({ ...qrOptions, [name]: value });
  };

  const handleDownload = async () => {
    if (qrOptions.logoImage && !isLogoLoaded) {
      alert("Please wait for the logo to load before downloading.");
      return;
    }

    const canvas = await html2canvas(qrRef.current!);
    canvas.toBlob((blob) => {
      if (blob) saveAs(blob, "qr-code.png");
    });
  };

  useEffect(() => {
    if (qrOptions.logoImage) {
      const img = new Image();
      img.src = user.logo;
      img.onload = () => setIsLogoLoaded(true);
      img.onerror = () => {
        alert("Failed to load logo image.");
        setIsLogoLoaded(false);
      };
    } else {
      setIsLogoLoaded(true);
    }
  }, [qrOptions.logoImage]);

  return (
    <div className={styles.createQrCode}>
      <h1>QR Code Generator</h1>
      <div className={styles.qrCodeGrid}>
        <div className={styles.editSection}>
          <div className={styles.qrTemplateSelection}>
            {QrCodeTemplates.map((template) => (
              <div
                key={template.name}
                className={`${styles.template} ${
                  template.name === qrOptions.name ? styles.selected : ""
                }`}
              >
                <QrCodeTemplate
                  {...template}
                  url={qrOptions.url}
                  logo={user.logo}
                  setQrOptions={() => {
                    setQrOptions({
                      ...qrOptions,
                      ...template,
                      size: 150,
                    });
                  }}
                />
              </div>
            ))}
          </div>

          <div className={styles.backgroundColorSelection}>
            <label>Background Color: </label>
            <input
              type="color"
              name="backgroundColor"
              value={qrOptions.backgroundColor}
              onChange={handleInputChange}
            />
          </div>

          <div className={styles.backgroundColorSelection}>
            <label>Qr Code Background Color: </label>
            <input
              type="color"
              name="bgColor"
              value={qrOptions.bgColor}
              onChange={handleInputChange}
            />
          </div>

          <div className={styles.textColorSelection}>
            <label>Color: </label>
            <input
              type="color"
              name="fgColor"
              value={qrOptions.fgColor}
              onChange={handleInputChange}
            />
          </div>

          <div className={styles.textColorSelection}>
            <label>Text Color: </label>
            <input
              type="color"
              name="textColor"
              value={qrOptions.textColor}
              onChange={handleInputChange}
            />
          </div>

          <div>
            <label>Logo Width: </label>
            <input
              type="number"
              name="logoWidth"
              value={qrOptions.logoWidth}
              onChange={handleInputChange}
            />
          </div>

          <div>
            <label>Logo Height: </label>
            <input
              type="number"
              name="logoHeight"
              value={qrOptions.logoHeight}
              onChange={handleInputChange}
            />
          </div>

          <div>
            <label>Eye Radius: </label>
            <input
              type="number"
              name="eyeRadius"
              value={qrOptions.eyeRadius}
              onChange={handleInputChange}
            />
          </div>

          <div className={styles.textColorSelection}>
            <label>Eye color: </label>
            <input
              type="color"
              name="eyeColor"
              value={qrOptions.eyeColor}
              onChange={handleInputChange}
            />
          </div>

          <div>
            <label>QR Style: </label>
            <select
              name="qrStyle"
              value={qrOptions.qrStyle}
              onChange={handleInputChange}
            >
              <option value="squares">Squares</option>
              <option value="dots">Dots</option>
              <option value="fluid">Fluid</option>
            </select>
          </div>

          <div>
            <label>title: </label>
            <input
              type="text"
              name="text"
              value={qrOptions.text}
              onChange={handleInputChange}
            />
          </div>

          <div>
            <label>Description: </label>
            <input
              type="text"
              name="description"
              value={qrOptions.description}
              onChange={handleInputChange}
            />
          </div>

          <div>
            <label>Display Logo: </label>
            <input
              type="checkbox"
              name="logoImage"
              checked={qrOptions.logoImage}
              onChange={handleInputChange}
            />
          </div>
        </div>

        <div className={styles.previewSection}>
          <div
            ref={qrRef}
            className={styles.qrCodeContainer}
            style={{
              width: qrOptions.size + 200,
              height: qrOptions.size + 200,
              backgroundColor: qrOptions.backgroundColor || qrOptions.bgColor,
            }}
          >
            {qrOptions.text && (
              <p
                style={{
                  color: qrOptions.textColor || qrOptions.fgColor,
                }}
              >
                <strong>{qrOptions.text}</strong>
              </p>
            )}
            <QRCode
              value={qrOptions.url}
              size={qrOptions.size + 100}
              fgColor={qrOptions.fgColor}
              bgColor={qrOptions.bgColor}
              logoImage={qrOptions.logoImage ? user?.logo : ""}
              logoWidth={qrOptions.logoWidth}
              logoHeight={qrOptions.logoHeight}
              eyeRadius={qrOptions.eyeRadius}
              qrStyle={qrOptions.qrStyle}
              eyeColor={qrOptions.eyeColor}
              logoPadding={5}
              removeQrCodeBehindLogo
            />
            {qrOptions.description && (
              <p
                style={{
                  color: qrOptions.textColor || qrOptions.fgColor,
                }}
              >
                {qrOptions.description}
              </p>
            )}
          </div>
          <button className={styles.downloadButton} onClick={handleDownload}>
            Download QR Code
          </button>
        </div>
      </div>
    </div>
  );
}

export default QrCode;

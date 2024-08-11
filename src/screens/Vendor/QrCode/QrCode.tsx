import { useState, useRef, useEffect } from "react";
import { QRCode } from "react-qrcode-logo";
import html2canvas from "html2canvas";
import { saveAs } from "file-saver";
import { useSelector } from "react-redux";
import { RootState } from "../../../redux/store";

interface QrOptions {
  text: string;
  size: number;
  fgColor: string;
  bgColor: string;
  logoImage: string;
  logoWidth: number;
  logoHeight: number;
  eyeRadius: number;
  qrStyle: "squares" | "dots" | "fluid" | undefined;
}

function QrCode() {
  const userId = useSelector((state: RootState) => state.user.userId);
  const [qrOptions, setQrOptions] = useState<QrOptions>({
    text: "https://vbitez.voxire.com/restaurant/" + userId,
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffffff",
    logoImage: "",
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
      img.src = qrOptions.logoImage;
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
    <div className="create-qr-code">
      <h1>QR Code Generator</h1>
      <p>{qrOptions.text}</p>

      <div className="qr-template-selection">
        {/* Template options should go here */}
        <div className="template">Template 1</div>
        <div className="template selected">Template 2</div>
        {/* Add more templates as necessary */}
      </div>

      <div className="background-color-selection">
        <label>Background Color: </label>
        <input
          type="color"
          name="bgColor"
          value={qrOptions.bgColor}
          onChange={handleInputChange}
        />
      </div>

      <div className="text-color-selection">
        <label>Text Color: </label>
        <input
          type="color"
          name="fgColor"
          value={qrOptions.fgColor}
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
      <div>
        <label>QR Style: </label>
        <select
          name="qrStyle"
          value={qrOptions.qrStyle}
          onChange={handleInputChange}
        >
          <option value="squares">Squares</option>
          <option value="dots">Dots</option>
          <option value="fluid">fluid</option>
        </select>
      </div>
      <div
        ref={qrRef}
        className="qr-code-container"
        style={{
          width: qrOptions.size,
          height: qrOptions.size,
        }}
      >
        <QRCode
          value={qrOptions.text}
          size={qrOptions.size}
          fgColor={qrOptions.fgColor}
          bgColor={qrOptions.bgColor}
          logoImage={qrOptions.logoImage}
          logoWidth={qrOptions.logoWidth}
          logoHeight={qrOptions.logoHeight}
          eyeRadius={qrOptions.eyeRadius}
          qrStyle={qrOptions.qrStyle}
          logoPadding={5}
          removeQrCodeBehindLogo
        />
      </div>
      <button className="download-button" onClick={handleDownload}>
        Download QR Code
      </button>
    </div>
  );
}

export default QrCode;

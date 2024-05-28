import { useState, useRef, useEffect } from "react";
import { QRCode } from "react-qrcode-logo";
import html2canvas from "html2canvas";
import { saveAs } from "file-saver";
import VbitezButton from "../../../components/Buttons/Buttons";
import { useSelector } from "react-redux";

function CreateQrCode() {
  const userId = useSelector((state) => state.user.userId);
  const [qrOptions, setQrOptions] = useState({
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
  const logoRef = useRef(null);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setQrOptions({ ...qrOptions, [name]: value });
  };

  const handleDownload = async () => {
    if (qrOptions.logoImage && !isLogoLoaded) {
      alert("Please wait for the logo to load before downloading.");
      return;
    }

    const canvas = await html2canvas(qrRef.current);
    canvas.toBlob((blob) => {
      saveAs(blob, "qr-code.png");
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
      <div>
        <label>Size: </label>
        <input
          type="number"
          name="size"
          value={qrOptions.size}
          onChange={handleInputChange}
        />
      </div>
      <div>
        <label>Foreground Color: </label>
        <input
          type="color"
          name="fgColor"
          value={qrOptions.fgColor}
          onChange={handleInputChange}
        />
      </div>
      <div>
        <label>Background Color: </label>
        <input
          type="color"
          name="bgColor"
          value={qrOptions.bgColor}
          onChange={handleInputChange}
        />
      </div>
      <div>
        <label>Logo URL: </label>
        <input
          type="text"
          name="logoImage"
          value={qrOptions.logoImage}
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
        </select>
      </div>
      <div
        ref={qrRef}
        style={{
          width: qrOptions.size,
          height: qrOptions.size,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          margin: "20px",
        }}
        className="qr-code-container"
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
          logoBackgroundColor="#ffffff"
          logoPadding={5}
          removeQrCodeBehindLogo
        />
      </div>
      <VbitezButton onClick={handleDownload}>Download QR Code</VbitezButton>
    </div>
  );
}

export default CreateQrCode;

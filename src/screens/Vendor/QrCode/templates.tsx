import { QRCode } from "react-qrcode-logo";

export interface QrOptions {
  name?: string;
  text?: string;
  description?: string;
  size: number;
  fgColor: string;
  bgColor: string;
  logoImage?: boolean;
  logoWidth: number;
  logoHeight: number;
  eyeRadius: number;
  qrStyle: "squares" | "dots" | "fluid" | undefined;
  backgroundColor?: string;
  url?: string;
  textColor?: string;
  eyeColor?: string;
}

export const QrCodeTemplates: QrOptions[] = [
  {
    name: "template1",
    text: "",
    description: "",
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffffff",
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: "squares",
    backgroundColor: "#ffffff",
  },
  {
    name: "template2",
    text: "Table",
    description: "Scan Me",
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffffff",
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: "squares",
  },
  {
    name: "template3",
    text: "Table",
    description: "Scan Me",
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffeb3b",
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: "squares",
  },
  {
    name: "template4",
    text: "Table",
    description: "Scan Me",
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffeb3b",
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: "squares",
  },
  {
    name: "template5",
    text: "Table",
    description: "Scan Me",
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffeb3b",
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: "squares",
  },
  {
    name: "template6",
    text: "Table",
    description: "Scan Me",
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffeb3b",
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: "squares",
  },
  {
    name: "template7",
    text: "Table",
    description: "Scan Me",
    size: 150,
    fgColor: "#000000",
    bgColor: "#ffffff",
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: "squares",
  },
];

interface QrCodeTemplateProps extends QrOptions {
  setQrOptions: () => void;
}

export const QrCodeTemplate: React.FC<QrCodeTemplateProps> = ({
  name,
  url,
  text,
  description,
  fgColor,
  bgColor,
  logoImage,
  logoWidth,
  logoHeight,
  eyeRadius,
  qrStyle,
  backgroundColor,
  setQrOptions,
}) => {
  return (
    <div key={name} className="qrOptions" onClick={() => setQrOptions()}>
      <div
        style={{
          width: 100,
          height: 100,
          backgroundColor: backgroundColor || bgColor,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        <QRCode
          value={url}
          size={50}
          fgColor={fgColor}
          bgColor={bgColor}
          logoWidth={logoWidth}
          logoHeight={logoHeight}
          eyeRadius={eyeRadius}
          qrStyle={qrStyle}
          logoPadding={5}
          removeQrCodeBehindLogo
        />
      </div>
    </div>
  );
};

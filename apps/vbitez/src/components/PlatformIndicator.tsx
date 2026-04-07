import React from "react";
import { usePlatform } from "../hooks/usePlatform";
import { Chip } from "@voxire/components";
import { Monitor, Globe } from "lucide-react";
import "./PlatformIndicator.scss";

const PlatformIndicator: React.FC = () => {
  const { isTauri, platform } = usePlatform();

  return (
    <div className="platform-indicator">
      <Chip
        icon={isTauri ? <Monitor size={12} /> : <Globe size={12} />}
        label={isTauri ? "Desktop App" : "Web App"}
        color={isTauri ? "primary" : "secondary"}
        size="sm"
      />
      <span className="platform-indicator__caption">{platform} BETA</span>
    </div>
  );
};

export default PlatformIndicator;

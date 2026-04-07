import React, { PropsWithChildren } from 'react';
import { FullScreenLoader } from '@voxire/components';
import { useFeatures } from '../hooks/useFeatures';
import { useFeaturesStore } from '../stores/featuresStore';

const FeaturesGate: React.FC<PropsWithChildren> = ({ children }) => {
  useFeatures();
  const { loading } = useFeaturesStore();

  if (loading) {
    return <FullScreenLoader />;
  }

  return <>{children}</>;
};

export default FeaturesGate;

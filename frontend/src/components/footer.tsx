import { useEffect, useState } from 'react';
import { getVersionInfo } from '../lib/version';

export const Footer = () => {
  const [versionInfo, setVersionInfo] = useState({
    version: 'dev',
    gitRevision: 'unknown',
  });

  useEffect(() => {
    getVersionInfo().then(setVersionInfo);
  }, []);

  const shortHash = versionInfo.gitRevision.substring(0, 7);

  return (
    <footer className="bg-primary text-onprimary mt-4 py-3 px-4">
      <div className="max-w-6xl mx-auto flex flex-row flex-wrap items-center justify-center gap-2 text-sm">
        <div className="flex items-center gap-2">
          <span className="text-display-light">powered by</span>
          <a
            href="https://stashsphere.com"
            target="_blank"
            rel="noopener noreferrer"
            className="font-semibold hover:text-secondary transition duration-300"
          >
            StashSphere
          </a>
        </div>
        <div className="hidden md:block text-display-light">|</div>
        <a
          href="https://github.com/stashsphere"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-secondary transition duration-300"
        >
          GitHub
        </a>
        <div className="hidden md:block text-display-light">|</div>
        <a
          href="https://docs.stashsphere.com"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-secondary transition duration-300"
        >
          Documentation
        </a>
        <div className="hidden md:block text-display-light">|</div>
        <span className="text-display-light font-mono text-xs">{shortHash}</span>
      </div>
    </footer>
  );
};

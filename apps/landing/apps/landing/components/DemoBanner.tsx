function DemoBanner() {
  return (
    <div
      id="demo"
      style={{
        zIndex: '1000',
      }}
    >
      This is a demo - Check Out Our
      <a
        style={{
          color: 'var(--primary)',
          textDecoration: 'underline',
          marginLeft: '5px',
          cursor: 'pointer',
        }}
        href="https://voxire.com"
        target="_blank"
        rel="noopener noreferrer"
      >
        Packages
      </a>
    </div>
  );
}

export default DemoBanner;

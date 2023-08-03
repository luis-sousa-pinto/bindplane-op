export const GraphGradient: React.FC = () => {
  return (
    <div style={{ position: "absolute", top: "-100px" }}>
      <svg width="40" height="20">
        {/* This makes a gradient background available to the custom edges */}
        <defs>
          <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#4abaeb" />
            <stop offset="50%" stopColor="#5dc4b6" />
            <stop offset="100%" stopColor="#4abaeb" />
          </linearGradient>
          <pattern
            id="pattern"
            x="0"
            y="0"
            width="40"
            height="20"
            patternUnits="userSpaceOnUse"
          >
            <rect x="0" y="0" width="40" height="50" fill="url(#gradient)" />
          </pattern>
        </defs>
      </svg>
    </div>
  );
};

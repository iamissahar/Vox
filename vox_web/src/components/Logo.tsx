import React from "react";

interface LogoProps {
  size?: number;
}

const Logo: React.FC<LogoProps> = ({ size = 24 }) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 32 32"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
    <rect width="32" height="32" rx="8" fill="#e8ff5e" />
    <path
      d="M8 10l4 8 4-8"
      stroke="#080808"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      d="M20 10v8M20 14h4"
      stroke="#080808"
      strokeWidth="2"
      strokeLinecap="round"
    />
  </svg>
);

export default Logo;

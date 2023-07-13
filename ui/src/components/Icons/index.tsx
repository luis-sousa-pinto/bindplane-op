import React from "react";

export type IconProps = React.ComponentProps<"svg">;

/**
 * These SVGs are all open source icons from Feather: https://feathericons.com/
 */

export const KeyboardArrowDownIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill="currentColor"
      stroke="currentColor"
      strokeWidth="0"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M7.41 8.59 12 13.17l4.59-4.58L18 10l-6 6-6-6 1.41-1.41z"></path>
    </svg>
  );
};
export const ArrowUpIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <circle cx="12" cy="12" r="10"></circle>
      <polyline points="16 12 12 8 8 12"></polyline>
      <line x1="12" y1="16" x2="12" y2="8"></line>
    </svg>
  );
};

export const CheckIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <polyline points="20 6 9 17 4 12"></polyline>
    </svg>
  );
};

export const ChevronDown: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <polyline points="6 9 12 15 18 9"></polyline>
    </svg>
  );
};

export const ChevronUp: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <polyline points="18 15 12 9 6 15"></polyline>
    </svg>
  );
};

export const ChevronRight: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <polyline points="9 18 15 12 9 6"></polyline>
    </svg>
  );
};

export const ChevronLeft: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <polyline points="15 18 9 12 15 6"></polyline>
    </svg>
  );
};

export const CopyIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
    </svg>
  );
};

export const EditIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
      <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
    </svg>
  );
};

export const EmailIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
      <polyline points="22,6 12,13 2,6"></polyline>
    </svg>
  );
};

export const EyeIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
      <circle cx="12" cy="12" r="3"></circle>
    </svg>
  );
};

export const EyeOffIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
      <line x1="1" y1="1" x2="23" y2="23"></line>
    </svg>
  );
};

export const AgentGridIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      fill="none"
      height="20"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      viewBox="0 0 24 24"
      width="20"
      {...props}
    >
      <rect height="7" width="7" x="3" y="3"></rect>
      <rect height="7" width="7" x="14" y="3"></rect>
      <rect height="7" width="7" x="14" y="14"></rect>
      <rect height="7" width="7" x="3" y="14"></rect>
    </svg>
  );
};

export const GridIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      fill="currentColor"
      stroke="currentColor"
      height="24"
      strokeLinecap="round"
      strokeLinejoin="round"
      viewBox="-1 -1 28 28"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path d="M3.69231 0C1.66226 0 0 1.66226 0 3.69231C0 5.72236 1.66226 7.38462 3.69231 7.38462C4.34495 7.38462 4.94712 7.19712 5.48077 6.89423L7.5 9.14423C6.86178 9.9375 6.46154 10.9147 6.46154 12C6.46154 13.006 6.80048 13.9219 7.35577 14.6827L5.74038 16.3269C5.73317 16.3233 5.71875 16.3305 5.71154 16.3269C5.1274 15.9375 4.44231 15.6923 3.69231 15.6923C1.66226 15.6923 0 17.3546 0 19.3846C0 21.4147 1.66226 23.0769 3.69231 23.0769C5.72236 23.0769 7.38462 21.4147 7.38462 19.3846C7.38462 18.7788 7.21514 18.2163 6.95192 17.7115L8.71154 15.9519C9.40745 16.3738 10.2115 16.6154 11.0769 16.6154C11.7332 16.6154 12.357 16.4639 12.9231 16.2115L13.9904 17.7404C13.3377 18.4075 12.9231 19.3089 12.9231 20.3077C12.9231 22.3377 14.5853 24 16.6154 24C18.6454 24 20.3077 22.3377 20.3077 20.3077C20.3077 18.2776 18.6454 16.6154 16.6154 16.6154C16.2548 16.6154 15.905 16.6911 15.5769 16.7885L14.4231 15.1442C15 14.5313 15.4291 13.774 15.6058 12.9231H16.7596C17.1707 14.5096 18.5986 15.6923 20.3077 15.6923C22.3377 15.6923 24 14.03 24 12C24 9.96995 22.3377 8.30769 20.3077 8.30769C18.5986 8.30769 17.1707 9.49038 16.7596 11.0769H15.6058C15.3858 10.0168 14.7764 9.08654 13.9615 8.42308L14.3654 7.32692C14.4988 7.34135 14.6322 7.38462 14.7692 7.38462C16.7993 7.38462 18.4615 5.72236 18.4615 3.69231C18.4615 1.66226 16.7993 0 14.7692 0C12.7392 0 11.0769 1.66226 11.0769 3.69231C11.0769 4.91827 11.7079 5.98918 12.6346 6.66346L12.2885 7.55769C11.899 7.44952 11.4988 7.38462 11.0769 7.38462C10.2909 7.38462 9.56611 7.60817 8.91346 7.96154L6.80769 5.65385C7.17188 5.08414 7.38462 4.41346 7.38462 3.69231C7.38462 1.66226 5.72236 0 3.69231 0ZM3.69231 1.84615C4.72356 1.84615 5.53846 2.66106 5.53846 3.69231C5.53846 4.20793 5.32212 4.65865 4.99038 4.99038C4.65865 5.32212 4.20793 5.53846 3.69231 5.53846C2.66106 5.53846 1.84615 4.72356 1.84615 3.69231C1.84615 2.66106 2.66106 1.84615 3.69231 1.84615ZM14.7692 1.84615C15.8005 1.84615 16.6154 2.66106 16.6154 3.69231C16.6154 4.72356 15.8005 5.53846 14.7692 5.53846C13.738 5.53846 12.9231 4.72356 12.9231 3.69231C12.9231 2.66106 13.738 1.84615 14.7692 1.84615ZM11.0769 9.23077C12.6058 9.23077 13.8462 10.4712 13.8462 12C13.8425 12.0397 13.8425 12.0757 13.8462 12.1154C13.7849 13.5865 12.5661 14.7692 11.0769 14.7692C10.3846 14.7692 9.74639 14.524 9.25962 14.1058C8.67188 13.5974 8.30769 12.8365 8.30769 12C8.30769 10.4712 9.54808 9.23077 11.0769 9.23077ZM20.3077 10.1538C21.3389 10.1538 22.1538 10.9688 22.1538 12C22.1538 13.0313 21.3389 13.8462 20.3077 13.8462C19.2764 13.8462 18.4615 13.0313 18.4615 12C18.4615 10.9688 19.2764 10.1538 20.3077 10.1538ZM3.69231 17.5385C4.72356 17.5385 5.53846 18.3534 5.53846 19.3846C5.53846 20.4159 4.72356 21.2308 3.69231 21.2308C2.66106 21.2308 1.84615 20.4159 1.84615 19.3846C1.84615 18.3534 2.66106 17.5385 3.69231 17.5385ZM16.6154 18.4615C17.6466 18.4615 18.4615 19.2764 18.4615 20.3077C18.4615 21.3389 17.6466 22.1538 16.6154 22.1538C15.5841 22.1538 14.7692 21.3389 14.7692 20.3077C14.7692 19.6947 15.0685 19.1719 15.5192 18.8365C15.6599 18.7897 15.7897 18.7103 15.8942 18.6058C15.905 18.6022 15.9123 18.5805 15.9231 18.5769C16.1358 18.494 16.3702 18.4615 16.6154 18.4615Z" />
    </svg>
  );
};

export const HelpCircleIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      fill="none"
      height="20"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      viewBox="0 0 24 24"
      width="20"
      {...props}
    >
      <circle cx="12" cy="12" r="10" stroke="currentColor"></circle>
      <path
        d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"
        stroke="currentColor"
      ></path>
      <line stroke="currentColor" x1="12" x2="12.01" y1="17" y2="17"></line>
    </svg>
  );
};

export const LogoutIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
      <polyline points="16 17 21 12 16 7"></polyline>
      <line x1="21" y1="12" x2="9" y2="12"></line>
    </svg>
  );
};

export const MessageSquareIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      fill="none"
      height="24"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
    </svg>
  );
};

export const MenuIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <line x1="3" y1="12" x2="21" y2="12"></line>
      <line x1="3" y1="6" x2="21" y2="6"></line>
      <line x1="3" y1="18" x2="21" y2="18"></line>
    </svg>
  );
};

export const MoreVertical: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <circle cx="12" cy="12" r="1"></circle>
      <circle cx="12" cy="5" r="1"></circle>
      <circle cx="12" cy="19" r="1"></circle>
    </svg>
  );
};

export const OtherDestinationsIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      version="1.1"
      width="96"
      height="48"
      viewBox="0 0 1024 512"
      stroke="currentColor"
      fill="currentColor"
      {...props}
    >
      <path
        strokeWidth="32.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        fill="none"
        d="M 516.3078 68.69225 L 987.4129 68.69225 L 987.4129 437.00138 L 516.3078 437.00138 Z "
      />
      <path
        stroke="currentColor"
        fill="currentColor"
        transform="matrix(1,0,0,-1,0,512)"
        d="M 26.61157 237.7464 L 348.5715 237.7464 C 360.4457 237.7464 370.0715 247.3723 370.0715 259.2464 C 370.0715 271.1206 360.4457 280.7464 348.5715 280.7464 L 26.61157 280.7464 C 14.73744 280.7464 5.111565 271.1206 5.111565 259.2464 C 5.111565 247.3723 14.73744 237.7464 26.61157 237.7464 Z M 341.4769 200.4725 L 483.3696 259.2464 L 341.4769 318.0203 Z M 341.4769 200.4725 "
      />
    </svg>
  );
};

export const PauseIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="rgb(165,165,165)"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="feather feather-pause"
      {...props}
    >
      <rect x="6" y="4" width="4" height="16"></rect>
      <rect x="14" y="4" width="4" height="16"></rect>
    </svg>
  );
};

export const PencilIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
    </svg>
  );
};

export const PlayIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="rgb(46,163,92)"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      className="feather feather-play"
      {...props}
    >
      <polygon points="5 3 19 12 5 21 5 3"></polygon>
    </svg>
  );
};

export const PlusCircleIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      className="feather feather-plus-circle"
      fill="none"
      height="24"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      viewBox="0 0 24 24"
      width="24"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <circle cx="12" cy="12" r="10"></circle>
      <line x1="12" x2="12" y1="8" y2="16"></line>
      <line x1="8" x2="16" y1="12" y2="12"></line>
    </svg>
  );
};

export const ProcessorIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 512 512"
      {...props}
    >
      <path d="M416 48v416c0 26.51-21.49 48-48 48H144c-26.51 0-48-21.49-48-48V48c0-26.51 21.49-48 48-48h224c26.51 0 48 21.49 48 48zm96 58v12a6 6 0 0 1-6 6h-18v6a6 6 0 0 1-6 6h-42V88h42a6 6 0 0 1 6 6v6h18a6 6 0 0 1 6 6zm0 96v12a6 6 0 0 1-6 6h-18v6a6 6 0 0 1-6 6h-42v-48h42a6 6 0 0 1 6 6v6h18a6 6 0 0 1 6 6zm0 96v12a6 6 0 0 1-6 6h-18v6a6 6 0 0 1-6 6h-42v-48h42a6 6 0 0 1 6 6v6h18a6 6 0 0 1 6 6zm0 96v12a6 6 0 0 1-6 6h-18v6a6 6 0 0 1-6 6h-42v-48h42a6 6 0 0 1 6 6v6h18a6 6 0 0 1 6 6zM30 376h42v48H30a6 6 0 0 1-6-6v-6H6a6 6 0 0 1-6-6v-12a6 6 0 0 1 6-6h18v-6a6 6 0 0 1 6-6zm0-96h42v48H30a6 6 0 0 1-6-6v-6H6a6 6 0 0 1-6-6v-12a6 6 0 0 1 6-6h18v-6a6 6 0 0 1 6-6zm0-96h42v48H30a6 6 0 0 1-6-6v-6H6a6 6 0 0 1-6-6v-12a6 6 0 0 1 6-6h18v-6a6 6 0 0 1 6-6zm0-96h42v48H30a6 6 0 0 1-6-6v-6H6a6 6 0 0 1-6-6v-12a6 6 0 0 1 6-6h18v-6a6 6 0 0 1 6-6z" />
    </svg>
  );
};

export const UploadCloudIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <polyline points="16 16 12 12 8 16"></polyline>
      <line x1="12" y1="12" x2="12" y2="21"></line>
      <path d="M20.39 18.39A5 5 0 0 0 18 9h-1.26A8 8 0 1 0 3 16.3"></path>
      <polyline points="16 16 12 12 8 16"></polyline>
    </svg>
  );
};

export const UsersIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
      <circle cx="9" cy="7" r="4"></circle>
      <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
      <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
    </svg>
  );
};

export const RadioIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <circle cx="12" cy="12" r="2"></circle>
      <path d="M16.24 7.76a6 6 0 0 1 0 8.49m-8.48-.01a6 6 0 0 1 0-8.49m11.31-2.82a10 10 0 0 1 0 14.14m-14.14 0a10 10 0 0 1 0-14.14"></path>
    </svg>
  );
};

export const RefreshIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <polyline points="23 4 23 10 17 10"></polyline>
      <polyline points="1 20 1 14 7 14"></polyline>
      <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
    </svg>
  );
};

export const SearchIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <circle cx="11" cy="11" r="8"></circle>
      <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
    </svg>
  );
};

export const SendIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <line x1="22" y1="2" x2="11" y2="13"></line>
      <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
    </svg>
  );
};

export const SettingsIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      fill="none"
      height="20"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="2"
      viewBox="0 0 24 24"
      width="20"
      {...props}
    >
      <circle cx="12" cy="12" r="3"></circle>
      <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
    </svg>
  );
};

export const SlackIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M14.5 10c-.83 0-1.5-.67-1.5-1.5v-5c0-.83.67-1.5 1.5-1.5s1.5.67 1.5 1.5v5c0 .83-.67 1.5-1.5 1.5z"></path>
      <path d="M20.5 10H19V8.5c0-.83.67-1.5 1.5-1.5s1.5.67 1.5 1.5-.67 1.5-1.5 1.5z"></path>
      <path d="M9.5 14c.83 0 1.5.67 1.5 1.5v5c0 .83-.67 1.5-1.5 1.5S8 21.33 8 20.5v-5c0-.83.67-1.5 1.5-1.5z"></path>
      <path d="M3.5 14H5v1.5c0 .83-.67 1.5-1.5 1.5S2 16.33 2 15.5 2.67 14 3.5 14z"></path>
      <path d="M14 14.5c0-.83.67-1.5 1.5-1.5h5c.83 0 1.5.67 1.5 1.5s-.67 1.5-1.5 1.5h-5c-.83 0-1.5-.67-1.5-1.5z"></path>
      <path d="M15.5 19H14v1.5c0 .83.67 1.5 1.5 1.5s1.5-.67 1.5-1.5-.67-1.5-1.5-1.5z"></path>
      <path d="M10 9.5C10 8.67 9.33 8 8.5 8h-5C2.67 8 2 8.67 2 9.5S2.67 11 3.5 11h5c.83 0 1.5-.67 1.5-1.5z"></path>
      <path d="M8.5 5H10V3.5C10 2.67 9.33 2 8.5 2S7 2.67 7 3.5 7.67 5 8.5 5z"></path>
    </svg>
  );
};

export const SlidersIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <line x1="4" y1="21" x2="4" y2="14"></line>
      <line x1="4" y1="10" x2="4" y2="3"></line>
      <line x1="12" y1="21" x2="12" y2="12"></line>
      <line x1="12" y1="8" x2="12" y2="3"></line>
      <line x1="20" y1="21" x2="20" y2="16"></line>
      <line x1="20" y1="12" x2="20" y2="3"></line>
      <line x1="1" y1="14" x2="7" y2="14"></line>
      <line x1="9" y1="8" x2="15" y2="8"></line>
      <line x1="17" y1="16" x2="23" y2="16"></line>
    </svg>
  );
};

export const SquareIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      fill="currentColor"
      stroke="currentColor"
      height="20"
      viewBox="0 -1 20 20"
      width="20"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path d="M23.4646 0.557089C23.1153 0.202015 22.6486 0.00419615 22.1505 6.66056e-05L22.1349 0C21.1153 0 20.2788 0.829639 20.2704 1.84937C20.2664 2.34745 20.4564 2.81735 20.8058 3.17243C21.1551 3.5275 21.6217 3.72532 22.1198 3.72945L22.1355 3.72951C23.1551 3.72951 23.9915 2.89988 23.9999 1.88014C24.004 1.38207 23.8139 0.912164 23.4646 0.557089Z" />
      <path d="M1.88014 10.1353L1.86449 10.1352C0.844823 10.1352 0.00845643 10.9649 6.41193e-05 11.9847C-0.00839479 13.0128 0.821178 13.8562 1.84937 13.8648H1.86495C2.88462 13.8648 3.72105 13.0351 3.72945 12.0154C3.7379 10.9872 2.90833 10.1438 1.88014 10.1353Z" />
      <path d="M8.89748 3.11794L8.88182 3.11787C7.86216 3.11787 7.02572 3.94751 7.0174 4.96725C7.00894 5.99544 7.83851 6.83886 8.8667 6.84732L8.88236 6.84739C9.90195 6.84739 10.7384 6.01775 10.7468 4.99802C10.7552 3.96976 9.92567 3.12633 8.89748 3.11794Z" />
      <path d="M13.9651 8.18556L13.9494 8.18549C12.9298 8.18549 12.0934 9.01513 12.085 10.0349C12.0766 11.0631 12.9062 11.9065 13.9344 11.9149L13.95 11.915C14.9696 11.915 15.806 11.0854 15.8144 10.0656C15.8229 9.03738 14.9932 8.19402 13.9651 8.18556Z" />
      <path d="M20.8068 2.68811L14.7791 8.7312L15.2668 9.2177L21.2945 3.17461L20.8068 2.68811Z" />
      <path d="M10.2006 5.8146L9.7135 6.30173L12.6332 9.22145L13.1203 8.73433L10.2006 5.8146Z" />
      <path d="M7.56358 5.81404L2.69617 10.6814L3.18329 11.1686L8.0507 6.30116L7.56358 5.81404Z" />
      <path d="M20.7637 7.0947C20.2114 7.0947 19.7637 7.54241 19.7637 8.0947V23C19.7637 23.5523 20.2114 24 20.7637 24H22.9999C23.5522 24 23.9999 23.5523 23.9999 23V8.0947C23.9999 7.54241 23.5522 7.0947 22.9999 7.0947H20.7637Z" />
      <path d="M14.1758 14.1893C13.6235 14.1893 13.1758 14.637 13.1758 15.1893V23C13.1758 23.5523 13.6235 24 14.1758 24H16.412C16.9643 24 17.412 23.5523 17.412 23V15.1893C17.412 14.637 16.9643 14.1893 16.412 14.1893H14.1758Z" />
      <path d="M7.58777 13.6826C7.03548 13.6826 6.58777 14.1303 6.58777 14.6826V23C6.58777 23.5523 7.03548 24 7.58777 24H9.82402C10.3763 24 10.824 23.5523 10.824 23V14.6826C10.824 14.1303 10.3763 13.6826 9.82402 13.6826H7.58777Z" />
      <path d="M1 16.2164C0.447716 16.2164 0 16.6641 0 17.2164V23C0 23.5523 0.447715 24 1 24H3.23626C3.78854 24 4.23625 23.5523 4.23626 23L4.23631 17.2164C4.23631 16.6641 3.7886 16.2164 3.23631 16.2164H1Z" />
    </svg>
  );
};

export const TrashIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <polyline points="3 6 5 6 21 6"></polyline>
      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
    </svg>
  );
};

export const UserXIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
      <circle cx="8.5" cy="7" r="4"></circle>
      <line x1="18" y1="8" x2="23" y2="13"></line>
      <line x1="23" y1="8" x2="18" y2="13"></line>
    </svg>
  );
};

export const XIcon: React.FC<IconProps> = (props) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <line x1="18" y1="6" x2="6" y2="18"></line>
      <line x1="6" y1="6" x2="18" y2="18"></line>
    </svg>
  );
};

/**
 * Miscellaneous (non-feather) icons.
 */

export const BusinessIcon: React.FC<IconProps> = (props) => {
  return (
    <svg width="40" height="40" viewBox="0 0 40 40" fill="none" {...props}>
      <g clipPath="url(#clip0_150_2266)">
        <path
          d="M23.3333 15.15L38.0167 18.0667C39.2 18.3167 40 19.3167 40 20.5V37.5C40 38.8833 38.8833 40 37.5 40H22.5C22.9667 40 23.3333 39.6333 23.3333 39.1667V38.3333H37.5C37.95 38.3333 38.3333 37.9667 38.3333 37.5V20.5C38.3333 20.1167 38.0667 19.7667 37.6833 19.6833L23.3333 16.85V15.15Z"
          fill="black"
        />
        <path
          d="M32.5 23.3333C32.9667 23.3333 33.3334 23.7 33.3334 24.1666C33.3334 24.6333 32.9667 25 32.5 25H29.1667C28.7 25 28.3334 24.6333 28.3334 24.1666C28.3334 23.7 28.7 23.3333 29.1667 23.3333H32.5Z"
          fill="black"
        />
        <path
          d="M32.5 28.3333C32.9667 28.3333 33.3334 28.7 33.3334 29.1666C33.3334 29.6333 32.9667 30 32.5 30H29.1667C28.7 30 28.3334 29.6333 28.3334 29.1666C28.3334 28.7 28.7 28.3333 29.1667 28.3333H32.5Z"
          fill="black"
        />
        <path
          d="M32.5 33.3333C32.9667 33.3333 33.3334 33.7 33.3334 34.1666C33.3334 34.6333 32.9667 35 32.5 35H29.1667C28.7 35 28.3334 34.6333 28.3334 34.1666C28.3334 33.7 28.7 33.3333 29.1667 33.3333H32.5Z"
          fill="black"
        />
        <path
          d="M23.3333 39.1667C23.3333 39.6333 22.9666 40 22.5 40C22.0333 40 21.6666 39.6333 21.6666 39.1667V38.3333V15.8333C21.6666 15.5833 21.7833 15.35 21.9666 15.1833C22.1666 15.0333 22.4166 14.9667 22.6666 15.0167L23.3333 15.15V16.85V38.3333V39.1667Z"
          fill="black"
        />
        <path
          d="M21.6667 38.3333V39.1666C21.6667 39.6333 22.0334 40 22.5 40H15.8334C16.3 40 16.6667 39.6333 16.6667 39.1666V38.3333H21.6667Z"
          fill="black"
        />
        <path
          d="M17.5 8.33331C17.9667 8.33331 18.3334 8.69998 18.3334 9.16665C18.3334 9.63331 17.9667 9.99998 17.5 9.99998H14.1667C13.7 9.99998 13.3334 9.63331 13.3334 9.16665C13.3334 8.69998 13.7 8.33331 14.1667 8.33331H17.5Z"
          fill="black"
        />
        <path
          d="M18.3334 14.1666C18.3334 14.6333 17.9667 15 17.5 15H14.1667C13.7 15 13.3334 14.6333 13.3334 14.1666C13.3334 13.7 13.7 13.3333 14.1667 13.3333H17.5C17.9667 13.3333 18.3334 13.7 18.3334 14.1666Z"
          fill="black"
        />
        <path
          d="M17.5 18.3333C17.9667 18.3333 18.3334 18.7 18.3334 19.1666C18.3334 19.6333 17.9667 20 17.5 20H14.1667C13.7 20 13.3334 19.6333 13.3334 19.1666C13.3334 18.7 13.7 18.3333 14.1667 18.3333H17.5Z"
          fill="black"
        />
        <path
          d="M17.5 23.3333C17.9667 23.3333 18.3334 23.7 18.3334 24.1666C18.3334 24.6333 17.9667 25 17.5 25H14.1667C13.7 25 13.3334 24.6333 13.3334 24.1666C13.3334 23.7 13.7 23.3333 14.1667 23.3333H17.5Z"
          fill="black"
        />
        <path
          d="M10 24.1666C10 24.6333 9.63333 25 9.16667 25H5.83333C5.36667 25 5 24.6333 5 24.1666C5 23.7 5.36667 23.3333 5.83333 23.3333H9.16667C9.63333 23.3333 10 23.7 10 24.1666Z"
          fill="black"
        />
        <path
          d="M9.16667 8.33331C9.63333 8.33331 10 8.69998 10 9.16665C10 9.63331 9.63333 9.99998 9.16667 9.99998H5.83333C5.36667 9.99998 5 9.63331 5 9.16665C5 8.69998 5.36667 8.33331 5.83333 8.33331H9.16667Z"
          fill="black"
        />
        <path
          d="M9.16667 13.3333C9.63333 13.3333 10 13.7 10 14.1666C10 14.6333 9.63333 15 9.16667 15H5.83333C5.36667 15 5 14.6333 5 14.1666C5 13.7 5.36667 13.3333 5.83333 13.3333H9.16667Z"
          fill="black"
        />
        <path
          d="M9.16667 18.3333C9.63333 18.3333 10 18.7 10 19.1666C10 19.6333 9.63333 20 9.16667 20H5.83333C5.36667 20 5 19.6333 5 19.1666C5 18.7 5.36667 18.3333 5.83333 18.3333H9.16667Z"
          fill="black"
        />
        <path
          d="M15 30.8333C15 30.3666 14.6166 30 14.1666 30H9.16663C8.69996 30 8.33329 30.3666 8.33329 30.8333V38.3333H6.66663V30.8333C6.66663 29.45 7.78329 28.3333 9.16663 28.3333H14.1666C15.55 28.3333 16.6666 29.45 16.6666 30.8333V38.3333H15V30.8333Z"
          fill="black"
        />
        <path
          d="M8.33329 38.3333H15H16.6666V39.1666C16.6666 39.6333 16.3 40 15.8333 40H7.49996C7.03329 40 6.66663 39.6333 6.66663 39.1666V38.3333H8.33329Z"
          fill="black"
        />
        <path
          d="M2.91667 0.333321L21.2333 3.11665C22.45 3.31665 23.3333 4.34999 23.3333 5.58332V15.15L22.6667 15.0167C22.4167 14.9667 22.1667 15.0333 21.9667 15.1833C21.7833 15.35 21.6667 15.5833 21.6667 15.8333V5.58332C21.6667 5.16665 21.3667 4.81665 20.9667 4.74999L2.65 1.98332C2.6 1.96665 2.55 1.96665 2.5 1.96665C2.3 1.96665 2.11667 2.03332 1.96667 2.16665C1.76667 2.33332 1.66667 2.54999 1.66667 2.79999V37.5C1.66667 37.9667 2.05 38.3333 2.5 38.3333H6.66667V39.1667C6.66667 39.6333 7.03333 40 7.5 40H2.5C1.11667 40 0 38.8833 0 37.5V2.79999C0 2.06665 0.316667 1.36665 0.883333 0.899988C1.45 0.416654 2.18333 0.216654 2.91667 0.333321V0.333321Z"
          fill="black"
        />
      </g>
      <defs>
        <clipPath id="clip0_150_2266">
          <rect width="40" height="40" fill="white" />
        </clipPath>
      </defs>
    </svg>
  );
};

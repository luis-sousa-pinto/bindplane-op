import { createTheme, ThemeOptions } from "@mui/material/styles";

const themeOptions: ThemeOptions = {
  palette: {
    primary: {
      main: "#4abaeb",
      dark: "#4abaeb",
    },
    secondary: {
      main: "#595a5c",
    },
    trace: {
      main: "#54b359",
      contrastText: "#fff",
    },
    debug: {
      main: "#54b399",
      contrastText: "#fff",
    },
    warning: {
      main: "#d68e57",
      contrastText: "#fff",
    },
    fatal: {
      main: "#82392b",
      contrastText: "#fff",
    },
    "action-blue": {
      main: "#4184E3",
      contrastText: "#fff",
    },
  },
  typography: {
    allVariants: {
      color: "#2f2f31",
    },
    fontFamily: "'Nunito Sans', sans-serif;",
    fontWeightBold: 600,
    button: {
      fontWeight: 700,
    },
  },

  components: {
    MuiPaper: {
      defaultProps: {
        variant: "outlined",
      },
    },
    MuiButton: {
      styleOverrides: {
        contained: {
          textTransform: "none",
          borderRadius: 20,
          boxShadow: "none",
        },
        outlined: {
          textTransform: "none",
          borderRadius: 20,
        },
        text: {
          textTransform: "none",
          borderRadius: 20,
        },
        containedPrimary: {
          color: "#ffffff",
        },
      },
    },
    MuiTab: {
      styleOverrides: {
        root: {
          textTransform: "none",
        },
      },
    },
    MuiLinearProgress: {
      styleOverrides: {
        colorPrimary: {
          backgroundColor: "#CDCED1",
          color: "#4184E3",
          height: 8,
          borderRadius: 4,
        },
        barColorPrimary: {
          backgroundColor: "#4184E3",
          borderRadius: 4,
        },
      },
    },
  },
};

export const theme = createTheme(themeOptions);

// Declarations to extend the theme and component interfaces
declare module "@mui/material/styles" {
  interface Palette {
    debug: Palette["primary"];
    trace: Palette["primary"];
    fatal: Palette["primary"];
    "action-blue": Palette["primary"];
  }

  // allow configuration using `createTheme`
  interface PaletteOptions {
    debug?: PaletteOptions["primary"];
    trace: PaletteOptions["primary"];
    fatal: PaletteOptions["primary"];
    "action-blue": PaletteOptions["primary"];
  }
}

// Update the Chip's color prop options
declare module "@mui/material/Chip" {
  interface ChipPropsColorOverrides {
    fatal: true;
    debug: true;
    trace: true;
  }
}

// Update the Button's color prop options
declare module "@mui/material/Button" {
  interface ButtonPropsColorOverrides {
    "action-blue": true;
  }
}

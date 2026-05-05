/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './app/**/*.{js,jsx,ts,tsx}',
    './src/**/*.{js,jsx,ts,tsx}',
  ],
  presets: [require('nativewind/preset')],
  theme: {
    extend: {
      colors: {
        background: '#121212',
        card: '#1a1a1a',
        sidebar: '#161616',
        primary: '#ffb700',
        'primary-fg': '#121212',
        foreground: '#e1e1e1',
        'muted-fg': '#888888',
        border: '#292929',
        input: '#292929',
        secondary: '#242424',
        destructive: '#cc3030',
        online: '#22c55e',
        offline: '#71717a',
      },
      fontFamily: {
        sans: ['Inter_400Regular', 'sans-serif'],
        'sans-medium': ['Inter_500Medium', 'sans-serif'],
        'sans-semibold': ['Inter_600SemiBold', 'sans-serif'],
        'sans-bold': ['Inter_700Bold', 'sans-serif'],
        'serif-italic': ['SourceSerif4_400Regular_Italic', 'serif'],
      },
      borderRadius: {
        sm: '4px',
        md: '6px',
        lg: '10px',
        full: '9999px',
      },
    },
  },
  plugins: [],
};

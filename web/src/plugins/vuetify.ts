import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

import { createVuetify } from 'vuetify'
import defaults from './defaults'

export const vuetify = createVuetify({
  defaults,
  theme: {
    defaultTheme: 'dark',
    themes: {
      dark: {
        dark: true,
        colors: {
          primary: '#B6F24A',
          secondary: '#6CCFF6',
          info: '#6CCFF6',
          accent: '#F2C14E',
          error: '#FF5C8A',
          warning: '#FFB454',
          success: '#7ED957',

          background: '#20231F',
          surface: '#2A2D27',
          'on-background': '#F4F1E8',
          'on-surface': '#F4F1E8',
          'on-primary': '#172006',

          'commute-fast': '#B6F24A',
          'commute-mid': '#F2C14E',
          'commute-slow': '#FF5C8A',

          'header-bar': '#181A17',
          'map-bg': '#171A16',
          'popup-overlay': '#1D201B',
          shadow: '#000000',
        },
      },
    },
  },
})

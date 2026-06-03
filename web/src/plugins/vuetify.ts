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
          primary: '#A6E22E',
          secondary: '#66D9EF',
          info: '#66D9EF',
          accent: '#AE81FF',
          error: '#F92672',
          warning: '#FD971F',
          success: '#A6E22E',

          background: '#272822',
          surface: '#2D2E27',
          'on-background': '#F8F8F2',
          'on-surface': '#F8F8F2',

          'commute-fast': '#A6E22E',
          'commute-mid': '#FD971F',
          'commute-slow': '#F92672',

          'header-bar': '#1E1F1C',
          'map-bg': '#1B1C18',
          'popup-overlay': '#1E1F1C',
          shadow: '#000000',
        },
      },
    },
  },
})

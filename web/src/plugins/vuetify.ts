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
          primary: '#472AB2',
          secondary: '#4AEAD8',
          success: '#2e7d32',
          warning: '#f9a825',
          error: '#c62828',
          'header-bar': '#262626',
          'map-bg': '#131516',
          'popup-overlay': '#14161C',
          shadow: '#000000',
        },
      },
    },
  },
})

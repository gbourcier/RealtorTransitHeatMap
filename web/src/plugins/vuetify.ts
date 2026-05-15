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
        },
      },
    },
  },
})

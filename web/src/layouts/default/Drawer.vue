<template>
  <v-navigation-drawer>
    <v-list>
      <template v-for="(item, index) in items" :key="index">
        <template v-if="item.subheader">
          <v-list-subheader>{{ item.subheader }}</v-list-subheader>
        </template>

        <template v-else>
          <v-list-item :to="item.to" :title="item.title" exact color="primary" variant="elevated" elevation="0"
            v-bind="item" />
        </template>
      </template>
    </v-list>
  </v-navigation-drawer>
</template>

<script lang="ts" setup>
const items = [
  { subheader: 'Browse' },
  { title: 'Listings', prependIcon: 'mdi-view-dashboard-outline', to: '/listings' },
  { subheader: 'Settings' },
  { title: 'Realtor Scraper', prependIcon: 'mdi-cog-transfer-outline', to: '/scraper' },
  { title: 'GTFS Data', prependIcon: 'mdi-train', to: '/transit' },
]
</script>

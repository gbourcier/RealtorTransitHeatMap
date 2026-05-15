<script setup lang="ts">
import SchedulerPanel from "../components/SchedulerPanel.vue";
import GtfsRefreshPanel from "../components/GtfsRefreshPanel.vue";
</script>

<template>
    <v-container fluid class="pa-2 pa-sm-6">
        <v-row>
            <v-col cols="12" md="4">
                <SchedulerPanel job-type="refresh_gtfs" />
            </v-col>
            <v-col cols="12" md="8">
                <GtfsRefreshPanel />
            </v-col>
        </v-row>
    </v-container>
</template>

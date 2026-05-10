<script setup lang="ts">
import SchedulerPanel from "../components/SchedulerPanel.vue";
import ScrapeJobsPanel from "../components/ScrapeJobsPanel.vue";
</script>

<template>
    <v-container fluid class="pa-6">
        <v-row>
            <v-col cols="12" md="4">
                <SchedulerPanel />
            </v-col>
            <v-col cols="12" md="8">
                <ScrapeJobsPanel />
            </v-col>
        </v-row>
    </v-container>
</template>

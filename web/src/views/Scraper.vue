<script setup lang="ts">
import { ref } from "vue";
import SchedulerPanel from "../components/SchedulerPanel.vue";
import ScrapeJobsPanel from "../components/ScrapeJobsPanel.vue";

const schedulesVersion = ref(0);
</script>

<template>
    <v-container fluid class="pa-2 pa-sm-6">
        <v-row>
            <v-col cols="12" md="4">
                <SchedulerPanel job-type="scrape_realtor" @changed="schedulesVersion++" />
            </v-col>
            <v-col cols="12" md="8">
                <ScrapeJobsPanel :reload-key="schedulesVersion" />
            </v-col>
        </v-row>
    </v-container>
</template>

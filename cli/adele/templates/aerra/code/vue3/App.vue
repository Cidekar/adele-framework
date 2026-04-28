<script setup lang="ts">
import { RouterView } from 'vue-router'
import AlertBanner from './components/AlertBanner.vue'

// The banner is dismissable by click and gets replaced when api.ts publishes
// a new message. We deliberately don't auto-clear on route changes so a
// "saved → redirected" success flash survives the navigation (e.g. profile
// password change → /login carries the "Password reset complete" message).
</script>

<template>
    <AlertBanner />
    <RouterView />
</template>

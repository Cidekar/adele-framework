<script setup lang="ts">
import { banner, clearBanner } from '../banner'
</script>

<template>
    <div v-if="banner.kind" class="alert" :class="{ 'alert--error': banner.kind === 'error' }" @click="clearBanner">
        {{ banner.message }}
    </div>
</template>

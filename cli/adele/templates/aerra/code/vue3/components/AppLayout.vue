<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { apiGet, apiPost } from '../api'

type DashboardUser = { firstName?: string; lastName?: string; email?: string }

const user = ref<DashboardUser | null>(null)
const router = useRouter()

onMounted(async () => {
    const res = await apiGet<DashboardUser>('/dashboard/home')
    if (res.ok && res.user) user.value = res.user
    else router.push('/login')
})

async function logout() {
    await apiPost('/logout', {})
    router.push('/login')
}
</script>

<template>
    <div v-if="user" class="min-h-screen flex">
        <aside class="w-60 bg-pink-1000 text-pink-50 p-4">
            <div class="advatar mb-4"><span class="name">{{ user.firstName ?? '?' }}</span></div>
            <nav class="flex flex-col gap-2">
                <RouterLink to="/dashboard/home">Home</RouterLink>
                <RouterLink to="/dashboard/profile">Profile</RouterLink>
                <button @click="logout" class="text-left">Logout</button>
            </nav>
        </aside>
        <main class="flex-1 p-8">
            <slot />
        </main>
    </div>
    <div v-else class="centered-vh">Loading...</div>
</template>

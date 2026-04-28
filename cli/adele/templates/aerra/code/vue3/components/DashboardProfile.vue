<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiGet, apiPost } from '../api'
import AppLayout from './AppLayout.vue'

type ProfileUser = { firstName?: string; lastName?: string; email?: string }

const router = useRouter()

const firstName = ref('')
const lastName = ref('')
const email = ref('')
const profileErrors = ref<Record<string, string>>({})
const profileMessage = ref('')
const profileSuccess = ref(false)
const profileSubmitting = ref(false)

const currentPassword = ref('')
const newPassword = ref('')
const verifyPassword = ref('')
const passwordErrors = ref<Record<string, string>>({})
const passwordMessage = ref('')
const passwordSubmitting = ref(false)

onMounted(async () => {
    const res = await apiGet<ProfileUser>('/dashboard/profile')
    if (res.ok && res.user) {
        firstName.value = res.user.firstName ?? ''
        lastName.value = res.user.lastName ?? ''
        email.value = res.user.email ?? ''
    } else if (res.ok === false) {
        router.push('/login')
    }
})

async function saveProfile() {
    profileSubmitting.value = true
    profileErrors.value = {}
    profileMessage.value = ''
    profileSuccess.value = false
    const res = await apiPost('/dashboard/profile', {
        name: `${firstName.value} ${lastName.value}`.trim(),
        email: email.value,
    })
    profileSubmitting.value = false
    if (res.ok) {
        profileSuccess.value = true
        if (res.message) profileMessage.value = res.message
        return
    }
    if (res.errors) profileErrors.value = res.errors
    if (res.message) profileMessage.value = res.message
}

async function changePassword() {
    passwordSubmitting.value = true
    passwordErrors.value = {}
    passwordMessage.value = ''
    const res = await apiPost('/dashboard/profile-password', {
        'current-password': currentPassword.value,
        password: newPassword.value,
        'verify-password': verifyPassword.value,
    })
    passwordSubmitting.value = false
    if (res.ok) {
        router.push(res.redirect ?? '/login')
        return
    }
    if (res.errors) passwordErrors.value = res.errors
    if (res.message) passwordMessage.value = res.message
}
</script>

<template>
    <AppLayout>
        <div class="space-y-8">
            <section class="card">
                <div class="card-header text-pink-50 text-2xl font-medium pb-4">Profile Information</div>
                <form @submit.prevent="saveProfile">
                    <div class="form-group mb-4">
                        <label for="firstName" class="block text-pink-50 mb-1">First Name</label>
                        <input id="firstName" v-model="firstName" type="text" autocomplete="given-name" class="w-full p-2 rounded" />
                        <div v-if="profileErrors.name" class="text-pink-50 text-sm mt-1">{{ profileErrors.name }}</div>
                    </div>
                    <div class="form-group mb-4">
                        <label for="lastName" class="block text-pink-50 mb-1">Last Name</label>
                        <input id="lastName" v-model="lastName" type="text" autocomplete="family-name" class="w-full p-2 rounded" />
                    </div>
                    <div class="form-group mb-4">
                        <label for="email" class="block text-pink-50 mb-1">Email</label>
                        <input id="email" v-model="email" type="email" autocomplete="email" required class="w-full p-2 rounded" />
                        <div v-if="profileErrors.email" class="text-pink-50 text-sm mt-1">{{ profileErrors.email }}</div>
                    </div>
                    <div class="text-right pt-2">
                        <button type="submit" :disabled="profileSubmitting" class="button">
                            {{ profileSubmitting ? 'Saving...' : 'Save' }}
                        </button>
                    </div>
                </form>
            </section>

            <section class="card">
                <div class="card-header text-pink-50 text-2xl font-medium pb-4">Password</div>
                <form @submit.prevent="changePassword">
                    <div class="form-group mb-4">
                        <label for="current-password" class="block text-pink-50 mb-1">Current Password</label>
                        <input id="current-password" v-model="currentPassword" type="password" autocomplete="current-password" required class="w-full p-2 rounded" />
                        <div v-if="passwordErrors['current-password']" class="text-pink-50 text-sm mt-1">{{ passwordErrors['current-password'] }}</div>
                    </div>
                    <div class="form-group mb-4">
                        <label for="new-password" class="block text-pink-50 mb-1">New Password</label>
                        <input id="new-password" v-model="newPassword" type="password" autocomplete="new-password" required class="w-full p-2 rounded" />
                        <div v-if="passwordErrors.password" class="text-pink-50 text-sm mt-1">{{ passwordErrors.password }}</div>
                    </div>
                    <div class="form-group mb-4">
                        <label for="verify-password" class="block text-pink-50 mb-1">Confirm Password</label>
                        <input id="verify-password" v-model="verifyPassword" type="password" autocomplete="new-password" required class="w-full p-2 rounded" />
                        <div v-if="passwordErrors['verify-password']" class="text-pink-50 text-sm mt-1">{{ passwordErrors['verify-password'] }}</div>
                    </div>
                    <div class="text-right pt-2">
                        <button type="submit" :disabled="passwordSubmitting" class="button">
                            {{ passwordSubmitting ? 'Saving...' : 'Save' }}
                        </button>
                    </div>
                </form>
            </section>
        </div>
    </AppLayout>
</template>

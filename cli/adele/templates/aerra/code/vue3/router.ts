import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import Home from './components/Home.vue'
import Login from './components/Login.vue'
import Registration from './components/Registration.vue'
import Forgot from './components/Forgot.vue'
import ResetPassword from './components/ResetPassword.vue'
import DashboardHome from './components/DashboardHome.vue'
import DashboardProfile from './components/DashboardProfile.vue'

const routes: RouteRecordRaw[] = [
    { path: '/', component: Home, meta: { layout: 'public' } },
    { path: '/login', component: Login, meta: { layout: 'public' } },
    { path: '/registration', component: Registration, meta: { layout: 'public' } },
    { path: '/forgot', component: Forgot, meta: { layout: 'public' } },
    { path: '/reset-password', component: ResetPassword, meta: { layout: 'public' } },
    { path: '/dashboard/home', component: DashboardHome, meta: { layout: 'private' } },
    { path: '/dashboard/profile', component: DashboardProfile, meta: { layout: 'private' } },
]

export const router = createRouter({
    history: createWebHistory(),
    routes,
})

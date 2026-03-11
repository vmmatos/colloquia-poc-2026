<script setup lang="ts">
definePageMeta({ middleware: 'auth' })

// Profile is now managed via the ProfilePanel overlay in the default layout.
// Navigating to /profile opens the panel from the sidebar.
// This page auto-opens the panel and redirects to home.

const showProfile = inject<Ref<boolean>>('showProfile')

onMounted(() => {
  if (showProfile) {
    showProfile.value = true
  }
  navigateTo('/')
})
</script>

<template>
  <div />
</template>

<template>
  <div>

    <nav class="flex gap-6 p-4 border-b">

      <NuxtLink to="/">Home</NuxtLink>
      <NuxtLink to="/stats">Stats</NuxtLink>
      <NuxtLink to="/history">History</NuxtLink>

    </nav>

    <main class="p-6">
      <slot />
    </main>

  </div>
</template>
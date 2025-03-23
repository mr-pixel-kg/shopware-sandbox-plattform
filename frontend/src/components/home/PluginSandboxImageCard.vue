<script>
export default {
  name: "PluginSandboxImageCard"
}

// TODO Zukünftige Komponente um Plugin-Sandboxen kompakter anzuzeigen

</script>

<template>
  <div class="max-w-7xl mx-auto bg-red-500 p-5">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold text-gray-900">Meine Sandboxen</h1>
    </div>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6" id="theme-grid">

      <template v-for="n in 4">
        <Card>
          <template #content>
            <div class="flex items-center space-x-4 relative">
              <span class="absolute -top-2 -right-6 pi pi-trash rounded-full p-2 hover:bg-zinc-200 hover:cursor-pointer"></span>
              <div class="w-24 h-24 ">
                <img src="/plugin1.jpg" alt="Titel" class="w-24 h-24 object-cover rounded-lg"/>
              </div>
              <div class="flex-1">
                <h2 class="text-lg font-medium mb-2">Sandbox #123</h2>
                <p class="text-sm text-gray-500 mb-4">Description</p>
                <span class="text-sm hover:underline">Details</span>

                <Button size="small" label="Details öffnen" severity="secondary" icon="pi pi-info-circle" />
              </div>
            </div>

          </template>
        </Card>
      </template>

    </div>
  </div>
</template>

<style scoped>

</style>
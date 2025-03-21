<script>
import ProgressSpinner from 'primevue/progressspinner';
import Card from "primevue/card";
import { GeneralStore } from "@/stores/generalStore";

export default {
  components: {
    Card,
    ProgressSpinner,
  },

  data() {
    return {
      isLoading: this.showLoadingScreen
    };
  },

  setup() {
    const store = GeneralStore();
    return {
      generalStore: store
    }
  },

};

</script>

<template>
    <div class="loadingscreen" v-if="generalStore.isLoading">
      <div class="absolute inset-0 flex justify-center items-center z-10">

        <Card style="width: 35rem; overflow: hidden">
          <template #header>
            <img alt="user header" src="/shopware-banner.jpg" />
          </template>
          <template #title>
            Sandbox wird erstellt...
          </template>
          <template #content>
            <div class="p-6 justify-center">
              <ProgressSpinner style="width: 50px; height: 50px" strokeWidth="8" fill="transparent" animationDuration=".5s" aria-label="Custom ProgressSpinner" />
            </div>
          </template>
        </Card>

      </div>
    </div>
</template>

<style scoped>
.loadingscreen {
  position: fixed;
  /*display: none;  Hidden by default */
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0,0,0,0.8);
  z-index: 2;
}
</style>
<template>
  <div class="upload">
    <div class="error">
      <p v-if="error"><span class="fa fa-warning"></span> {{ error }}</p>
      <span v-if="scanning">
                  <span class="fa fa-fw fa-spinner fa-spin"></span> Scanning...
      </span>
    </div>
  
    <dropzone id="myVueDropzone"
              :url="uploadUrl"
              maxFileSizeInMB="64"
              useFontAwesome="true"
              maxNumberOfFiles="64"
              showRemoveLink="false"
              v-bind:headers="headers"></dropzone>
  </div>
</template>

<script lang="ts">
    
    import Vue from 'vue';
    
    import * as $ from 'jquery';
    import * as _ from 'lodash';
    import Api from './../Api';
    import Song from './../Song';
    import Playlist from './../Playlist';
    import PubSub from './../PubSub';
    import AudioPlayerEvents from './../AudioPlayer';
    import AudioPlayer from './../AudioPlayer';

        interface UploadComponent extends Vue {
                    error: string;
                    scanning: boolean;
    }

export default {
  name: 'upload',
  data() {
    return {
      error: '',
      scanning: false,
      uploadUrl: '/api/upload',
    }
  },
  computed: {
        headers() {
              return {
                  Authorization: 'Bearer '+ Api.retrieveToken()
              } 
        } 
  },
  mounted: function() {
  },
  methods: {
  }

} as Vue.ComponentOptions<UploadComponent>;

</script>
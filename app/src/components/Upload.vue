<template>
  <div class="upload">
    <div class="error">
      <p v-if="error"><span class="fa fa-warning"></span> {{ error }}</p>
      <span v-if="scanning">
          <span class="fa fa-fw fa-spinner fa-spin"></span> Scanning...
      </span>
    </div>
    
    <dropzone id="myVueDropzone" url="/api/upload" maxFileSizeInMB="64" useFontAwesome="true" maxNumberOfFiles="64" showRemoveLink="false" v-bind:headers="headers"></dropzone>

  </div>
</template>

<script>
    
//import Auth from '../Auth'
import {Promise} from 'es6-promise'; // for Travis
let Api = require('./../Api').default;
//let router = require('./../main').router;

export default {
  data() {
    return {
      error: '',
      scanning: false,
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

}
</script>
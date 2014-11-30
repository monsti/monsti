(function() {
  $(document).ready(function () {
    tinymce.init({
      selector: ".html-field textarea",
      plugins: "anchor autosave code hr image visualchars visualblocks table paste media link",
      tools: "inserttable",
      height: 300,
    });
  });
})();

(function() {
  var disableEditor = function(textarea) {
    var editable = textarea.previousSibling;
    $(editable).mahalo();
    textarea.value = editable.innerHTML;
    editable.parentElement.removeChild(editable);
    textarea.style.display = "block";
    textarea.attributes.editorActive = false;
  }

  var enableEditor = function(textarea) {
    textarea.style.display = "none";
    var editable = document.createElement("div");
    editable.classList.add("editable");
    editable.innerHTML = textarea.value;
    textarea.parentElement.insertBefore(editable, textarea);
    Aloha.jQuery(editable).aloha();
    textarea.attributes.editorActive = true;
  }

  // AddEditor transforms the given textarea into an editable for Aloha.
  AddEditor = function(textarea) {
    var button = document.createElement("button");
    button.type = "button";
    button.appendChild(document.createTextNode("Toggle raw HTML"));
    button.addEventListener("click", function() {
      if (textarea.attributes.editorActive == true) {
        disableEditor(textarea);
      } else {
        enableEditor(textarea);
      }
    });
    textarea.parentElement.insertBefore(button, textarea.nextSibling)
    textarea.form.addEventListener("submit", function() {
      if (textarea.attributes.editorActive == true) {
        disableEditor(textarea);
      }
    });
    enableEditor(textarea);
  }

  $(document).ready(function () {
    Aloha.ready( function() {
      $('.html-field textarea').each(function() { AddEditor(this); });
    });
  });

})();

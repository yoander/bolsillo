$(document).ready(function () {
  $('input.flexdatalist').on('change:flexdatalist after:flexdatalist.remove', function(event, set) {
    var tags = undefined;
    var input = undefined;
    if (event.type == 'change:flexdatalist') {
      if (set.value != undefined) {
        var input = $('input#tags');
        var tags = JSON.parse(input.val());
        var id = $('datalist#tag-list option[data-tag="' + set.value + '"]').val();
        tags.push(id);
      }
    } else if (event.type == 'after:flexdatalist') {
      var input = $('input#tags');
      var tags = JSON.parse(input.val());
      var id = $('datalist#tag-list option[data-tag="' + $(set).children(':first').text() + '"]').val();
      tags.splice(tags.indexOf(id), 1);
    }
    if ((tags != undefined) && (input != undefined)) {
      input.val(JSON.stringify(tags));
    }
  })
});

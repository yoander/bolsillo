$(document).ready(function () {
    $('.tag').click(function () {
        var tag = $(this);
        var id = tag.attr('id')
        console.log("ID", id)
        var input = $('input#tags');
        var tags = JSON.parse(input.val());
        if (tag.hasClass('badge-danger')) {
            tags.push(id);
            tag.removeClass('badge-danger').addClass('badge-info').html('✔' + tag.html());
        } else if (tag.hasClass('badge-info')) {
            tags.splice(tags.indexOf(id), 1);
            tag.removeClass('badge-info').addClass('badge-danger').html(tag.html().replace('✔', ''));
        }
        input.val(JSON.stringify(tags));
    });
});
/*
$(document).on('click', '#dropdownMenuButton span a', function (e) {
    console.log('PASO');
    console.log(e.which);
});
*/

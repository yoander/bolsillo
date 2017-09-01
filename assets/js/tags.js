$('#d ul li').click(function (e) {
    $(this).children('span').removeClass('badge-secondary').addClass('badge-info');
    $('#dropdownMenuButton').html($('#dropdownMenuButton').html() + " " + $(this).html());
    e.stopPropagation();
});
/*
$('#d').on('show.bs.dropdown', function () {
    

    $('#dropdownMenuButton span').each(function (index, ele) {
        $('#d ul li span').each(function (j, ea) {
            console.log($(ele).text() + '==' + $(ea).text());
            if ( $(ele).text() == $(ea).text()) {
                $(ea).parent().remove();  
            }
        });
    });   
})*/
$(document).ready(function () {
    $('#filter').click(function() {
        var filter = $(this);
        var faIcon = $(this).children('i');
        $.ajax({
            url: $(this).attr('data-url'),
            data: {
               startDate: $('#startDate').val(),
               endDate: $('#endDate').val(),
               keyword: $('#description').val(),
            },
            beforeSend: function() {
                filter.prop('disabled', true);
                faIcon.removeClass('fa-filter')
                    .addClass('fa-spinner fa-spin fa-fw');    
            },
            success: function(data) {
                $('#transTable').html(data);
            },
            complete: function () {
                faIcon.removeClass('fa-spinner fa-spin fa-fw')
                    .addClass('fa-filter');
                filter.prop('disabled', false);
                feather.replace();
            }
        });
    }) ;
});
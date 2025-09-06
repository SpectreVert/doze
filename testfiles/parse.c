int parse(char *str)
{
    int ret = 0;
    while (*(str + ret))
    {
        ret++;
    }
    return ret;
}
